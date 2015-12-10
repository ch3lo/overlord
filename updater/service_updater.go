package updater

import (
	"reflect"
	"sync"
	"time"

	"github.com/ch3lo/overlord/manager/cluster"
	"github.com/ch3lo/overlord/scheduler"
	"github.com/ch3lo/overlord/util"
)

const service_updated string = "updated"
const service_added string = "added"
const service_removed string = "removed"

type ServiceUpdaterData struct {
	lastUpdate time.Time
	lastAction string
	origin     scheduler.ServiceInformation
}

type ServiceUpdater struct {
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

		for clusterKey, c := range su.clusters {
			srvs, err := c.GetScheduler().GetInstances(scheduler.FilterInstances{})
			if err != nil {
				util.Log.Errorf("No se pudieron obtener instancias del cluster %s con scheduler %s. Motivo: %s", clusterKey, c.GetScheduler().Id(), err.Error())
				continue
			}
			checkedServices := su.checkServices(srvs)
			for k := range checkedServices {
				updatedServices[k] = checkedServices[k]
			}
		}

		util.Log.Debugf("%v", updatedServices)

		if updatedServices != nil && len(updatedServices) > 0 {
			su.notify(updatedServices)
		}

		time.Sleep(time.Second * 30)
	}
}

func (su *ServiceUpdater) checkServices(services []scheduler.ServiceInformation) map[string]*ServiceUpdaterData {
	updatedServices := make(map[string]*ServiceUpdaterData)

	// Se asume por defecto que un servicio fue removido
	// Luego se actualiza al estado correcto
	for k := range su.services {
		su.services[k].lastAction = service_removed
		su.services[k].lastUpdate = time.Now()
	}

	for k, v := range services {
		util.Log.Debugf("Comparando servicio %+v <-> %+v", su.services[v.Id], services[k])

		if su.services[v.Id] == nil {
			newService := &ServiceUpdaterData{
				lastAction: service_added,
				lastUpdate: time.Now(),
				origin:     services[k],
			}

			su.services[v.Id] = newService
			updatedServices[v.Id] = newService

			util.Log.Debugln("Monitoreando nuevo servicio")
			continue
		}

		su.services[v.Id].lastUpdate = time.Now()
		su.services[v.Id].lastAction = service_updated
		if reflect.DeepEqual(su.services[v.Id].origin, services[k]) {
			util.Log.Debugln("Servicio sin cambios")
			continue
		}

		su.services[v.Id].origin = services[k]
		updatedServices[v.Id] = su.services[v.Id]

		util.Log.Debugln("Servicio tuvo un cambio")
	}

	return updatedServices
}
