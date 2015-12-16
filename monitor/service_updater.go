package monitor

import (
	"reflect"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ch3lo/overlord/cluster"
	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/scheduler"
	"github.com/ch3lo/overlord/util"
)

type ServiceDataStatus int

const (
	SERVICE_UPDATED ServiceDataStatus = 1 + iota
	SERVICE_ADDED
	SERVICE_REMOVED
	SERVICE_UPDATING
)

var statuses = [...]string{
	"SERVICE_UPDATED",
	"SERVICE_ADDED",
	"SERVICE_REMOVED",
	"SERVICE_UPDATING",
}

func (s ServiceDataStatus) String() string {
	return statuses[s-1]
}

type ServiceUpdaterData struct {
	registerDate time.Time
	lastUpdate   time.Time
	lastAction   ServiceDataStatus
	clusterId    string
	origin       scheduler.ServiceInformation
}

func NewServiceUpdaterData() *ServiceUpdaterData {
	data := &ServiceUpdaterData{
		registerDate: time.Now(),
		lastAction:   SERVICE_ADDED,
		lastUpdate:   time.Now(),
	}
	return data
}

func (data *ServiceUpdaterData) RegisterDate() time.Time              { return data.registerDate }
func (data *ServiceUpdaterData) LastUpdate() time.Time                { return data.lastUpdate }
func (data *ServiceUpdaterData) LastAction() ServiceDataStatus        { return data.lastAction }
func (data *ServiceUpdaterData) ClusterId() string                    { return data.clusterId }
func (data *ServiceUpdaterData) Origin() scheduler.ServiceInformation { return data.origin }
func (data *ServiceUpdaterData) InStatus(status ServiceDataStatus) bool {
	return data.lastAction == status
}

type ServiceUpdater struct {
	updateServicesMux  sync.Mutex
	subscriberMux      sync.Mutex
	interval           time.Duration
	subscribers        map[string]ServiceUpdaterSubscriber
	subscriberCriteria map[string]ServiceChangeCriteria
	clusters           map[string]*cluster.Cluster
	services           map[string]*ServiceUpdaterData
}

func NewServiceUpdater(config configuration.Updater, clusters map[string]*cluster.Cluster) *ServiceUpdater {
	if clusters == nil || len(clusters) == 0 {
		util.Log.Fatalln("Al menos se debe monitorear un cluster")
	}

	interval := time.Second * 10
	if config.Interval != 0 {
		interval = config.Interval
	}

	s := &ServiceUpdater{
		interval:           interval,
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

	if _, ok := su.subscribers[sub.Id()]; ok {
		return
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

	if _, ok := su.subscribers[sub.Id()]; ok {
		delete(su.subscriberCriteria, sub.Id())
		delete(su.subscribers, sub.Id())
		util.Log.Infof("Se removio el subscriptor: %s", sub.Id())
		return
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
			srvs, err := c.GetScheduler().Instances(scheduler.FilterInstances{})
			if err != nil {
				util.Log.WithFields(log.Fields{
					"cluster":   clusterKey,
					"scheduler": c.GetScheduler().ID(),
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

		time.Sleep(su.interval)
	}
}

func (su *ServiceUpdater) checkClusterServices(clusterId string, clusterServices []scheduler.ServiceInformation) map[string]*ServiceUpdaterData {
	updatedServices := make(map[string]*ServiceUpdaterData)

	// Se asume por defecto que un servicio esta actualizandose
	// Luego se actualiza al estado correcto
	// Si el servicio ya fue removido no se toma en cuenta
	// Si se remueve como falso positivo, se volvera a agregar el servicio al map en la iteracion
	// pero se interpretara como un servicio nuevo
	for k := range su.services {
		if su.services[k].clusterId == clusterId && su.services[k].lastAction != SERVICE_REMOVED {
			su.services[k].lastAction = SERVICE_UPDATING
			su.services[k].lastUpdate = time.Now()
			updatedServices[k] = su.services[k]
		}
	}

	for k, v := range clusterServices {
		if su.services[v.ID] == nil {
			newService := NewServiceUpdaterData()
			newService.clusterId = clusterId
			newService.origin = clusterServices[k]

			su.services[v.ID] = newService
			updatedServices[v.ID] = newService

			util.Log.WithField("cluster", clusterId).Debugf("Monitoreando nuevo servicio %+v", newService)
			continue
		}

		su.services[v.ID].lastUpdate = time.Now()
		su.services[v.ID].lastAction = SERVICE_UPDATED

		if reflect.DeepEqual(su.services[v.ID].origin, clusterServices[k]) {
			delete(updatedServices, v.ID)
			util.Log.WithField("cluster", clusterId).Debugf("Servicio sin cambios %+v", clusterServices[k])
			continue
		}

		su.services[v.ID].origin = clusterServices[k]
		util.Log.WithField("cluster", clusterId).Debugf("Servicio tuvo un cambio %+v <-> %+v", su.services[v.ID].Origin(), clusterServices[k])
	}

	for k := range su.services {
		if su.services[k].lastAction == SERVICE_UPDATING {
			util.Log.WithField("cluster", clusterId).Debugf("Servicio removido %+v", su.services[k])
			su.services[k].lastAction = SERVICE_REMOVED
			su.services[k].lastUpdate = time.Now()
		}
	}

	return updatedServices
}
