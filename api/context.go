package api

import (
	"fmt"
	"sync"

	"github.com/ch3lo/overlord/cluster"
	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/logger"
	"github.com/ch3lo/overlord/manager/report"
	"github.com/ch3lo/overlord/manager/service"
	"github.com/ch3lo/overlord/monitor"
	"github.com/ch3lo/overlord/notification"
	"github.com/ch3lo/overlord/notification/factory"
	//Necesarios para que funcione el init()
	_ "github.com/ch3lo/overlord/notification/email"
	_ "github.com/ch3lo/overlord/notification/http"
	_ "github.com/latam-airlines/mesos-framework-factory/marathon"
	_ "github.com/latam-airlines/mesos-framework-factory/swarm"
)

type appContext struct {
	serviceMux         sync.Mutex
	config             *configuration.Configuration
	serviceUpdater     *monitor.ServiceUpdater
	broadcaster        report.Broadcast
	clusters           map[string]*cluster.Cluster
	serviceGroupMapper map[string]*service.Application
}

func newContext(config *configuration.Configuration) *appContext {
	app := &appContext{
		config:             config,
		clusters:           make(map[string]*cluster.Cluster),
		serviceGroupMapper: make(map[string]*service.Application),
	}

	app.setupBroadcaster(config.Notification)
	app.setupClusters(config.Clusters)
	app.setupServiceUpdater(config.Updater)

	return app
}

// setupBroadcaster inicializa el broadcaster de Notificaciones
func (o *appContext) setupBroadcaster(config configuration.Notification) {
	broadcaster := report.NewBroadcaster(config.AttemptsOnError, config.WaitOnError, config.WaitAfterAttemts)
	var notifications []notification.Notification
	for key, params := range config.Providers {
		if params.Disabled {
			logger.Instance().Warnf("El notificador no esta habilitado: %s", key)
			continue
		}

		notification, err := factory.Create(params.NotificationType, key, params.Config)
		if err != nil {
			logger.Instance().Fatalf("Error al crear la notificacion %s. %s", key, err.Error())
		}

		logger.Instance().Infof("Se creo nuevo notificador %s de tipo %s", key, params.NotificationType)
		notifications = append(notifications, notification)
		logger.Instance().Infof("Registrando el notificador %s en broadcaster", key)
		if err := broadcaster.Register(notification); err != nil {
			logger.Instance().Fatalf("No se pudo registrar el notificador en el broadcaster: %s", err.Error())
		}
	}

	if len(notifications) == 0 {
		logger.Instance().Warnln("No hay notificadores configurados")
	}

	o.broadcaster = broadcaster
}

// setupClusters inicia el cluster, mapeando el cluster el id del cluster como key
func (o *appContext) setupClusters(config map[string]configuration.Cluster) {
	for key := range config {
		c, err := cluster.NewCluster(key, config[key])
		if err != nil {
			switch err.(type) {
			case *cluster.ClusterDisabled:
				logger.Instance().Warnln(err.Error())
				continue
			default:
				logger.Instance().Fatalln(err.Error())
			}
		}

		o.clusters[key] = c
		logger.Instance().Infof("Se configuro el cluster %s", key)
	}

	if len(o.clusters) == 0 {
		logger.Instance().Fatalln("Al menos debe existir un cluster")
	}
}

// setupServiceUpdater inicia el componente que monitorea cambios de servicios
func (o *appContext) setupServiceUpdater(config configuration.Updater) {
	su := monitor.NewServiceUpdater(config, o.clusters)
	su.Monitor()
	o.serviceUpdater = su
}

func (o *appContext) clusterIds() []string {
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
func (o *appContext) RegisterService(params service.Parameters) (*service.Manager, error) {
	o.serviceMux.Lock()
	defer o.serviceMux.Unlock()

	group := o.RegisterApplication(params)

	sm, err := group.RegisterServiceManager(o.clusterIds(), o.config.Manager.Check, o.broadcaster, params)
	if err != nil {
		return nil, err
	}

	imageNameRegexp, _ := sm.FullImageNameRegexp()
	criteria := &monitor.ImageNameAndImageTagRegexpCriteria{imageNameRegexp}
	o.serviceUpdater.Register(sm, criteria)
	return sm, nil
}

func (o *appContext) RegisterApplication(params service.Parameters) *service.Application {
	var group *service.Application
	var ok bool
	if group, ok = o.serviceGroupMapper[params.ID]; ok {
		group = o.serviceGroupMapper[params.ID]
	} else {
		group = service.NewServiceGroup(params.ID)
		o.serviceGroupMapper[params.ID] = group
	}
	return group
}

func (o *appContext) GetApplications() map[string]*service.Application {
	return o.serviceGroupMapper
}

// NotificationDisabled error generado cuando un notificador no esta habilitado
type NotificationDisabled struct {
	Name string
}

func (err NotificationDisabled) Error() string {
	return fmt.Sprintf("El notificador no esta habilitado: %s", err.Name)
}
