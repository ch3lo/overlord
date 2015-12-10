package manager

import (
	"sync"

	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/manager/cluster"
	"github.com/ch3lo/overlord/service"
	"github.com/ch3lo/overlord/updater"
	"github.com/ch3lo/overlord/util"
	//Necesarios para que funcione el init()
	_ "github.com/ch3lo/overlord/scheduler/marathon"
	_ "github.com/ch3lo/overlord/scheduler/swarm"
)

var overlordApp *Overlord

type Overlord struct {
	config           *configuration.Configuration
	serviceMux       sync.Mutex
	clusters         map[string]*cluster.Cluster
	serviceUpdater   *updater.ServiceUpdater
	serviceContainer map[string]*service.ServiceContainer
}

func NewApp(config *configuration.Configuration) {
	app := &Overlord{
		config:           config,
		serviceContainer: make(map[string]*service.ServiceContainer),
	}

	app.setupClusters(config)
	app.setupServiceUpdater()

	overlordApp = app
}

func GetAppInstance() *Overlord {
	return overlordApp
}

// setupClusters inicia el cluster, mapeando el cluster el id del cluster como key
func (o *Overlord) setupClusters(config *configuration.Configuration) {
	o.clusters = make(map[string]*cluster.Cluster)

	for key, _ := range config.Clusters {
		c, err := cluster.NewCluster(key, config.Clusters[key])
		if err != nil {
			util.Log.Infof(err.Error())
			continue
		}

		o.clusters[key] = c
	}

	if len(o.clusters) == 0 {
		util.Log.Fatalln("Al menos debe existir un cluster")
	}
}

// setupServiceUpdater inicia el componente que monitorea cambios de servicios
func (o *Overlord) setupServiceUpdater() {
	su := updater.NewServiceUpdater(o.clusters)
	su.Monitor()
	o.serviceUpdater = su
}

// RegisterService registra un nuevo servicio en un contenedor de servicios
// Si el contenedor ya existia se omite su creación y se procede a registrar
// las versiones de los servicios.
// Si no se puede registrar una nueva version se retornara un error.
func (o *Overlord) RegisterService(params service.ServiceParameters) (*service.ServiceVersion, error) {
	o.serviceMux.Lock()
	defer o.serviceMux.Unlock()

	container := service.NewServiceContainer(params.Id)

	for key, _ := range o.serviceContainer {
		if key == params.Id {
			container = o.serviceContainer[key]
		}
	}

	o.serviceContainer[params.Id] = container

	sv, err := container.RegisterServiceVersion(params)
	if err != nil {
		return nil, err
	}

	criteria := &updater.ImageNameCriteria{sv.ImageName + ":" + sv.ImageTag}

	o.serviceUpdater.Register(sv, criteria)
	return sv, nil
}

func (o *Overlord) GetServices() map[string]*service.ServiceContainer {
	return o.serviceContainer
}
