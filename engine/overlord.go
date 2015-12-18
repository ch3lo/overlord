package engine

import (
	"fmt"
	"sync"

	"github.com/ch3lo/overlord/cluster"
	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/manager/report"
	"github.com/ch3lo/overlord/manager/service"
	"github.com/ch3lo/overlord/monitor"
	"github.com/ch3lo/overlord/notification"
	"github.com/ch3lo/overlord/notification/factory"
	"github.com/ch3lo/overlord/util"
	//Necesarios para que funcione el init()
	_ "github.com/ch3lo/overlord/notification/email"
	_ "github.com/ch3lo/overlord/notification/http"
	_ "github.com/ch3lo/overlord/scheduler/marathon"
	_ "github.com/ch3lo/overlord/scheduler/swarm"
)

var overlordApp *Overlord

type Overlord struct {
	serviceMux         sync.Mutex
	config             *configuration.Configuration
	serviceUpdater     *monitor.ServiceUpdater
	broadcaster        report.Broadcast
	clusters           map[string]*cluster.Cluster
	serviceGroupMapper map[string]*service.Group
}

func NewApp(config *configuration.Configuration) {
	app := &Overlord{
		config:             config,
		clusters:           make(map[string]*cluster.Cluster),
		serviceGroupMapper: make(map[string]*service.Group),
	}

	app.setupBroadcaster(config.Notification)
	app.setupClusters(config.Clusters)
	app.setupServiceUpdater(config.Updater)

	overlordApp = app
}

func GetAppInstance() *Overlord {
	return overlordApp
}

// setupBroadcaster inicializa el broadcaster de Notificaciones
func (o *Overlord) setupBroadcaster(config configuration.Notification) {
	broadcaster := report.NewBroadcaster(config.AttemptsOnError, config.WaitOnError, config.WaitAfterAttemts)
	var notifications []notification.Notification
	for key, params := range config.Providers {
		if params.Disabled {
			util.Log.Warnf("El notificador no esta habilitado: %s", key)
			continue
		}

		notification, err := factory.Create(params.NotificationType, key, params.Config)
		if err != nil {
			util.Log.Fatalf("Error al crear la notificacion %s. %s", key, err.Error())
		}

		util.Log.Infof("Se creo nuevo notificador %s de tipo %s", key, params.NotificationType)
		notifications = append(notifications, notification)
		util.Log.Infof("Registrando el notificador %s en broadcaster", key)
		if err := broadcaster.Register(notification); err != nil {
			util.Log.Fatalf("No se pudo registrar el notificador en el broadcaster: %s", err.Error())
		}
	}

	if len(notifications) == 0 {
		util.Log.Warnln("No hay notificadores configurados")
	}

	o.broadcaster = broadcaster

}

// setupClusters inicia el cluster, mapeando el cluster el id del cluster como key
func (o *Overlord) setupClusters(config map[string]configuration.Cluster) {
	for key := range config {
		c, err := cluster.NewCluster(key, config[key])
		if err != nil {
			switch err.(type) {
			case *cluster.ClusterDisabled:
				util.Log.Warnln(err.Error())
				continue
			default:
				util.Log.Fatalln(err.Error())
			}
		}

		o.clusters[key] = c
		util.Log.Infof("Se configuro el cluster %s", key)
	}

	if len(o.clusters) == 0 {
		util.Log.Fatalln("Al menos debe existir un cluster")
	}
}

// setupServiceUpdater inicia el componente que monitorea cambios de servicios
func (o *Overlord) setupServiceUpdater(config configuration.Updater) {
	su := monitor.NewServiceUpdater(config, o.clusters)
	su.Monitor()
	o.serviceUpdater = su
}

func (o *Overlord) clusterIds() []string {
	var names []string
	for k := range o.clusters {
		names = append(names, k)
	}
	return names
}

// RegisterService registra un nuevo servicio en un contenedor de servicios
// Si el contenedor ya existia se omite su creación y se procede a registrar
// las versiones de los servicios.
// Si no se puede registrar una nueva version se retornara un error.
func (o *Overlord) RegisterService(params service.Parameters) (*service.Manager, error) {
	o.serviceMux.Lock()
	defer o.serviceMux.Unlock()

	group := o.RegisterGroup(params)

	sm, err := group.RegisterServiceManager(o.clusterIds(), o.config.Manager.Check, o.broadcaster, params)
	if err != nil {
		return nil, err
	}

	imageNameRegexp, _ := sm.FullImageNameRegexp()
	criteria := &monitor.ImageNameAndImageTagRegexpCriteria{imageNameRegexp}
	o.serviceUpdater.Register(sm, criteria)
	return sm, nil
}

func (o *Overlord) RegisterGroup(params service.Parameters) *service.Group {
	var group *service.Group
	var ok bool
	if group, ok = o.serviceGroupMapper[params.ID]; ok {
		group = o.serviceGroupMapper[params.ID]
	} else {
		group = service.NewServiceGroup(params.ID)
		o.serviceGroupMapper[params.ID] = group
	}
	return group
}

func (o *Overlord) GetServices() map[string]*service.Group {
	return o.serviceGroupMapper
}

// NotificationDisabled error generado cuando un notificador no esta habilitado
type NotificationDisabled struct {
	Name string
}

func (err NotificationDisabled) Error() string {
	return fmt.Sprintf("El notificador no esta habilitado: %s", err.Name)
}
