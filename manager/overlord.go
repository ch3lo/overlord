package manager

import (
	"sync"

	"github.com/ch3lo/overlord/cluster"
	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/manager/service"
	"github.com/ch3lo/overlord/monitor"
	"github.com/ch3lo/overlord/util"
	//Necesarios para que funcione el init()
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
}

func NewApp(config *configuration.Configuration) {
	app := &Overlord{
		config:             config,
		serviceGroupMapper: make(map[string]*service.ServiceGroup),
	}

	clusters := setupClusters(config)

	for k := range clusters {
		app.clusterNames = append(app.clusterNames, k)
	}

	app.setupServiceUpdater(config, clusters)

	overlordApp = app
}

func GetAppInstance() *Overlord {
	return overlordApp
}

// setupClusters inicia el cluster, mapeando el cluster el id del cluster como key
func setupClusters(config *configuration.Configuration) map[string]*cluster.Cluster {
	clusters := make(map[string]*cluster.Cluster)

	for key, _ := range config.Clusters {
		c, err := cluster.NewCluster(key, config.Clusters[key])
		if err != nil {
			util.Log.Infof(err.Error())
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
func (o *Overlord) setupServiceUpdater(config *configuration.Configuration, clusters map[string]*cluster.Cluster) {
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
	if o.groupExists(params.Id) {
		group = o.serviceGroupMapper[params.Id]
	} else {
		group = service.NewServiceGroup(params.Id)
		o.serviceGroupMapper[params.Id] = group
	}
	return group
}

func (o *Overlord) groupExists(groupId string) bool {
	groupExists := false
	for key, _ := range o.serviceGroupMapper {
		if key == groupId {
			groupExists = true
			break
		}
	}

	return groupExists
}

func (o *Overlord) GetServices() map[string]*service.ServiceGroup {
	return o.serviceGroupMapper
}
