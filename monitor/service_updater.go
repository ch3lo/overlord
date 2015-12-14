package monitor

import (
	"reflect"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ch3lo/overlord/cluster"
	"github.com/ch3lo/overlord/scheduler"
	"github.com/ch3lo/overlord/util"
)

const service_updated string = "updated"
const service_added string = "added"
const service_removed string = "removed"

type ServiceUpdaterData struct {
	registerDate time.Time
	lastUpdate   time.Time
	lastAction   string
	clusterId    string
	origin       scheduler.ServiceInformation
}

func NewServiceUpdaterData() *ServiceUpdaterData {
	data := &ServiceUpdaterData{
		registerDate: time.Now(),
		lastAction:   service_added,
		lastUpdate:   time.Now(),
	}
	return data
}

func (data *ServiceUpdaterData) GetRegisterDate() time.Time {
	return data.registerDate
}

func (data *ServiceUpdaterData) GetLastUpdate() time.Time {
	return data.lastUpdate
}

func (data *ServiceUpdaterData) GetLastAction() string {
	return data.lastAction
}

func (data *ServiceUpdaterData) GetClusterId() string {
	return data.clusterId
}

func (data *ServiceUpdaterData) GetOrigin() scheduler.ServiceInformation {
	return data.origin
}

type ServiceUpdater struct {
	updateServicesMux  sync.Mutex
	subscriberMux      sync.Mutex
	subscribers        map[string]ServiceUpdaterSubscriber
	subscriberCriteria map[string]ServiceChangeCriteria
	clusters           map[string]*cluster.Cluster
	services           map[string]*ServiceUpdaterData
}

func NewServiceUpdater(clusters map[string]*cluster.Cluster) *ServiceUpdater {
	if clusters == nil || len(clusters) == 0 {
		util.Log.Fatalln("Al menos se debe monitorear un cluster")
	}

	s := &ServiceUpdater{
		subscribers:        make(map[string]ServiceUpdaterSubscriber),
		subscriberCriteria: make(map[string]ServiceChangeCriteria),
		services:           make(map[string]*ServiceUpdaterData),
	}
	s.clusters = clusters

	return s
}

func (su *ServiceUpdater) Register(sub ServiceUpdaterSubscriber, cc ServiceChangeCriteria) {
	su.subscriberMux.Lock()
	defer su.subscriberMux.Unlock()

	for subscriberId, _ := range su.subscribers {
		if subscriberId == sub.Id() {
			return
		}
	}

	su.subscriberCriteria[sub.Id()] = cc
	su.subscribers[sub.Id()] = sub

	util.Log.Infof("Se agregó el subscriptor: %s", sub.Id())

	filtered := cc.MeetCriteria(su.services)
	if filtered != nil && len(filtered) > 0 {
		sub.Update(filtered)
	}
}

func (su *ServiceUpdater) Remove(sub ServiceUpdaterSubscriber) {
	su.subscriberMux.Lock()
	defer su.subscriberMux.Unlock()

	for k, v := range su.subscribers {
		if v.Id() == sub.Id() {
			delete(su.subscriberCriteria, k)
			delete(su.subscribers, k)
			util.Log.Infof("Se removio el subscriptor: %s", sub.Id())
			return
		}
	}
}

func (su *ServiceUpdater) notify(updatedServices map[string]*ServiceUpdaterData) {
	su.subscriberMux.Lock()
	defer su.subscriberMux.Unlock()

	util.Log.Debugln("Filtrando resultados y notificando cambios")
	for k, _ := range su.subscribers {
		filtered := su.subscriberCriteria[k].MeetCriteria(updatedServices)
		if filtered != nil && len(filtered) > 0 {
			su.subscribers[k].Update(filtered)
		}
	}
}

// Monitor comienza el monitoreo de los servicios de manera desatachada
func (su *ServiceUpdater) Monitor() {
	go su.detachedMonitor()
}

// detachedMonitor loop que permite monitorear los servicios de los schedulers
func (su *ServiceUpdater) detachedMonitor() {
	for {
		updatedServices := make(map[string]*ServiceUpdaterData)

		su.updateServicesMux.Lock()

		for clusterKey, c := range su.clusters {
			util.Log.WithField("cluster", clusterKey).Infof("Monitoreando cluster")
			srvs, err := c.GetScheduler().GetInstances(scheduler.FilterInstances{})
			if err != nil {
				util.Log.WithFields(log.Fields{
					"cluster":   clusterKey,
					"scheduler": c.GetScheduler().Id(),
				}).Errorf("No se pudieron obtener instancias del cluster. Motivo: %s", err.Error())
				continue
			}
			checkedServices := su.checkClusterServices(clusterKey, srvs)
			for k := range checkedServices {
				updatedServices[k] = checkedServices[k]
			}
			util.Log.WithField("cluster", clusterKey).Infof("Se actualizaron %d servicios", len(checkedServices))
		}

		su.updateServicesMux.Unlock()

		util.Log.Infof("Se actualizaron %d servicios", len(updatedServices))

		if updatedServices != nil && len(updatedServices) > 0 {
			su.notify(updatedServices)
		}

		time.Sleep(time.Second * 10)
	}
}

func (su *ServiceUpdater) checkClusterServices(clusterId string, clusterServices []scheduler.ServiceInformation) map[string]*ServiceUpdaterData {
	updatedServices := make(map[string]*ServiceUpdaterData)

	// Se asume por defecto que un servicio fue removido
	// Luego se actualiza al estado correcto
	for k := range su.services {
		if su.services[k].clusterId == clusterId {
			su.services[k].lastAction = service_removed
			su.services[k].lastUpdate = time.Now()
			updatedServices[k] = su.services[k]
		}
	}

	for k, v := range clusterServices {
		if su.services[v.Id] == nil {
			newService := NewServiceUpdaterData()
			newService.clusterId = clusterId
			newService.origin = clusterServices[k]

			su.services[v.Id] = newService
			updatedServices[v.Id] = newService

			util.Log.WithField("cluster", clusterId).Debugf("Monitoreando nuevo servicio %+v", newService)
			continue
		}

		su.services[v.Id].lastUpdate = time.Now()
		su.services[v.Id].lastAction = service_updated

		if reflect.DeepEqual(su.services[v.Id].origin, clusterServices[k]) {
			delete(updatedServices, v.Id)
			util.Log.WithField("cluster", clusterId).Debugf("Servicio sin cambios %+v", clusterServices[k])
			continue
		}

		su.services[v.Id].origin = clusterServices[k]
		util.Log.WithField("cluster", clusterId).Debugf("Servicio tuvo un cambio %+v <-> %+v", su.services[v.Id].GetOrigin(), clusterServices[k])
	}

	return updatedServices
}
