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

// ServiceDataStatus es un atributo de ServiceUpdaterData para especificar
// el estado en el que se encuentra el servicio
type ServiceDataStatus int

const (
	// ServiceUpdated estado de un servicio que se encuentra actualizado
	ServiceUpdated ServiceDataStatus = 1 + iota
	// ServiceAdded estado de un servicio que se encuentra recien agregado
	ServiceAdded
	// ServiceRemoved estado de un servicio que se encuentra recientemente removido
	ServiceRemoved
	// ServiceUpdating estado de un servicio que se encuentra en estado de actualizacion
	ServiceUpdating
)

var statuses = [...]string{
	"ServiceUpdated",
	"ServiceAdded",
	"ServiceRemoved",
	"ServiceUpdating",
}

func (s ServiceDataStatus) String() string {
	return statuses[s-1]
}

// ServiceUpdaterData basicamente es un decorador de scheduler.ServiceInformation
// que busca encapsular esta informacion y agregarle metadata de su estado de actualizacion
type ServiceUpdaterData struct {
	registerDate time.Time
	lastUpdate   time.Time
	lastAction   ServiceDataStatus
	clusterID    string
	origin       scheduler.ServiceInformation
}

// NewServiceUpdaterData crea una nueva instancia de ServiceUpdaterData
func NewServiceUpdaterData() *ServiceUpdaterData {
	data := &ServiceUpdaterData{
		registerDate: time.Now(),
		lastAction:   ServiceAdded,
		lastUpdate:   time.Now(),
	}
	return data
}

// RegisterDate obtiene la fecha de se creo esta instancia
func (data *ServiceUpdaterData) RegisterDate() time.Time { return data.registerDate }

// LastUpdate obtiene la fecha de cuando se actualizo por ultima vez esta instancia
func (data *ServiceUpdaterData) LastUpdate() time.Time { return data.lastUpdate }

// LastAction obtiene la ultima acción que se ejecuto sobre esta instancia
func (data *ServiceUpdaterData) LastAction() ServiceDataStatus { return data.lastAction }

// ClusterID obtiene el identificador del cluster al cual pertenece esta instancia
func (data *ServiceUpdaterData) ClusterID() string { return data.clusterID }

// Origin es un wrapper a la información obtenida desde el scheduler
func (data *ServiceUpdaterData) Origin() scheduler.ServiceInformation { return data.origin }

// InStatus retorna un bool tru si la instancia se encuentra en el estado pasado como parametro
func (data *ServiceUpdaterData) InStatus(status ServiceDataStatus) bool {
	return data.lastAction == status
}

// Serviceupdater es una estructura que maneja toda la información extraida de los cluster
// Acepta que se registren observers con un criterio de filtro.
// Para que estos observers sean notificados cuando se cumple con el criterio entregado
type ServiceUpdater struct {
	updateServicesMux  sync.Mutex
	subscriberMux      sync.Mutex
	interval           time.Duration
	subscribers        map[string]ServiceUpdaterSubscriber
	subscriberCriteria map[string]ServiceChangeCriteria
	clusters           map[string]*cluster.Cluster
	services           map[string]*ServiceUpdaterData
}

// NewServiceUpdater crea una nueva instancia de ServiceUpdater
// Necesita de al menos un cluster, sino se detendrá la ejecucion con un Fatal
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

// Register registra un nuevo observer/subscriptor con un criterio de filtro
// Cada vez que se obtiene informacion de los schedulers se le notificara al subsriptor de regreso
// con los servicios actualizados
func (su *ServiceUpdater) Register(sub ServiceUpdaterSubscriber, cc ServiceChangeCriteria) {
	su.subscriberMux.Lock()
	defer su.subscriberMux.Unlock()

	if _, ok := su.subscribers[sub.ID()]; ok {
		return
	}

	su.subscriberCriteria[sub.ID()] = cc
	su.subscribers[sub.ID()] = sub

	util.Log.Infof("Se agregó el subscriptor: %s", sub.ID())

	filtered := cc.MeetCriteria(su.services)
	if filtered != nil && len(filtered) > 0 {
		sub.Update(filtered)
	}
}

// Remove remueve un subscriptor
// TODO Remover los channels y detener funciones go de las dependencias
func (su *ServiceUpdater) Remove(sub ServiceUpdaterSubscriber) {
	su.subscriberMux.Lock()
	defer su.subscriberMux.Unlock()

	if _, ok := su.subscribers[sub.ID()]; ok {
		delete(su.subscriberCriteria, sub.ID())
		delete(su.subscribers, sub.ID())
		util.Log.Infof("Se removio el subscriptor: %s", sub.ID())
		return
	}
}

// Notify aplica los filtros de criterio sobre los updatedServices y notifica a los subscriptores
// si despues de aplicar el filtro existen resultados
func (su *ServiceUpdater) notify(updatedServices map[string]*ServiceUpdaterData) {
	su.subscriberMux.Lock()
	defer su.subscriberMux.Unlock()

	util.Log.Debugln("Filtrando resultados y notificando cambios")
	for k := range su.subscribers {
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

func (su *ServiceUpdater) checkClusterServices(clusterID string, clusterServices []scheduler.ServiceInformation) map[string]*ServiceUpdaterData {
	updatedServices := make(map[string]*ServiceUpdaterData)

	// Se asume por defecto que un servicio esta actualizandose
	// Luego se actualiza al estado correcto
	// Si el servicio ya fue removido no se toma en cuenta
	// Si se remueve como falso positivo, se volvera a agregar el servicio al map en la iteracion
	// pero se interpretara como un servicio nuevo
	for k := range su.services {
		if su.services[k].clusterID == clusterID && su.services[k].lastAction != ServiceRemoved {
			su.services[k].lastAction = ServiceUpdating
			su.services[k].lastUpdate = time.Now()
			updatedServices[k] = su.services[k]
		}
	}

	for k, v := range clusterServices {
		if su.services[v.ID] == nil {
			newService := NewServiceUpdaterData()
			newService.clusterID = clusterID
			newService.origin = clusterServices[k]

			su.services[v.ID] = newService
			updatedServices[v.ID] = newService

			util.Log.WithField("cluster", clusterID).Debugf("Monitoreando nuevo servicio %+v", newService)
			continue
		}

		su.services[v.ID].lastUpdate = time.Now()
		su.services[v.ID].lastAction = ServiceUpdated

		if reflect.DeepEqual(su.services[v.ID].origin, clusterServices[k]) {
			delete(updatedServices, v.ID)
			util.Log.WithField("cluster", clusterID).Debugf("Servicio sin cambios %+v", clusterServices[k])
			continue
		}

		su.services[v.ID].origin = clusterServices[k]
		util.Log.WithField("cluster", clusterID).Debugf("Servicio tuvo un cambio %+v <-> %+v", su.services[v.ID].Origin(), clusterServices[k])
	}

	for k := range su.services {
		if su.services[k].lastAction == ServiceUpdating {
			util.Log.WithField("cluster", clusterID).Debugf("Servicio removido %+v", su.services[k])
			su.services[k].lastAction = ServiceRemoved
			su.services[k].lastUpdate = time.Now()
		}
	}

	return updatedServices
}
