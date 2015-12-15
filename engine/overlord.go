package engine

import (
	"fmt"
	"sync"

	"github.com/ch3lo/overlord/cluster"
	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/manager/service"
	"github.com/ch3lo/overlord/monitor"
	"github.com/ch3lo/overlord/notification"
	"github.com/ch3lo/overlord/notification/factory"
	"github.com/ch3lo/overlord/util"
	//Necesarios para que funcione el init()
	_ "github.com/ch3lo/overlord/notification/email"
	_ "github.com/ch3lo/overlord/scheduler/marathon"
	_ "github.com/ch3lo/overlord/scheduler/swarm"
)

var overlordApp *Overlord

type Overlord struct {
	config             *configuration.Configuration
	clusterNames       []string
	serviceMux         sync.Mutex
	serviceUpdater     *monitor.ServiceUpdater
	serviceGroupMapper map[string]*service.ServiceGroup
	notifications      map[string]notification.Notification
}

func NewApp(config *configuration.Configuration) {
	app := &Overlord{
		config:             config,
		serviceGroupMapper: make(map[string]*service.ServiceGroup),
		notifications:      make(map[string]notification.Notification),
	}

	app.setupNotification(config.Notifications)

	clusters := setupClusters(config.Clusters)

	for k := range clusters {
		app.clusterNames = append(app.clusterNames, k)
	}

	app.setupServiceUpdater(config.Updater, clusters)

	overlordApp = app
}

func GetAppInstance() *Overlord {
	return overlordApp
}

// setupNotification inicializa los componentes de notificacion
func (o *Overlord) setupNotification(config map[string]configuration.Notification) {
	for key, params := range config {
		if params.Disabled {
			util.Log.Warnf("El notificador no esta habilitado: %s", key)
			continue
		}

		notification, err := factory.Create(params.NotificationType, params.Config)
		if err != nil {
			util.Log.Fatalf("Error al crear la notificacion %s. %s", key, err.Error())
		}

		util.Log.Infof("Se creo nuevo notificador %s de tipo %s", key, params.NotificationType)
		o.notifications[key] = notification
	}
}

// setupClusters inicia el cluster, mapeando el cluster el id del cluster como key
func setupClusters(config map[string]configuration.Cluster) map[string]*cluster.Cluster {
	clusters := make(map[string]*cluster.Cluster)

	for key, _ := range config {
		c, err := cluster.NewCluster(key, config[key])
		if err != nil {
			util.Log.Warnln(err.Error())
			continue
		}

		clusters[key] = c
	}

	if len(clusters) == 0 {
		util.Log.Fatalln("Al menos debe existir un cluster")
	}

	return clusters
}

// setupServiceUpdater inicia el componente que monitorea cambios de servicios
func (o *Overlord) setupServiceUpdater(config *configuration.Updater, clusters map[string]*cluster.Cluster) {
	su := monitor.NewServiceUpdater(config, clusters)
	su.Monitor()
	o.serviceUpdater = su
}

// RegisterService registra un nuevo servicio en un contenedor de servicios
// Si el contenedor ya existia se omite su creación y se procede a registrar
// las versiones de los servicios.
// Si no se puede registrar una nueva version se retornara un error.
func (o *Overlord) RegisterService(params service.ServiceParameters) (*service.ServiceManager, error) {
	o.serviceMux.Lock()
	defer o.serviceMux.Unlock()

	group := o.RegisterGroup(params)

	sm, err := group.RegisterServiceManager(o.clusterNames, params)
	if err != nil {
		return nil, err
	}

	imageNameRegexp, _ := sm.FullImageNameRegexp()
	criteria := &monitor.ImageNameAndImageTagRegexpCriteria{imageNameRegexp}
	o.serviceUpdater.Register(sm, criteria)
	return sm, nil
}

func (o *Overlord) RegisterGroup(params service.ServiceParameters) *service.ServiceGroup {
	var group *service.ServiceGroup
	var ok bool
	if group, ok = o.serviceGroupMapper[params.Id]; ok {
		group = o.serviceGroupMapper[params.Id]
	} else {
		group = service.NewServiceGroup(params.Id)
		o.serviceGroupMapper[params.Id] = group
	}
	return group
}

func (o *Overlord) GetServices() map[string]*service.ServiceGroup {
	return o.serviceGroupMapper
}

// NotificationDisabled error generado cuando un notificador no esta habilitado
type NotificationDisabled struct {
	Name string
}

func (err NotificationDisabled) Error() string {
	return fmt.Sprintf("El notificador no esta habilitado: %s", err.Name)
}
