package cluster

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/scheduler"
	"github.com/ch3lo/overlord/scheduler/factory"
	"github.com/ch3lo/overlord/util"
)

type Cluster struct {
	name      string
	scheduler scheduler.Scheduler
}

// NewCluster crea un nuevo cluster a partir de un id y parametros de configuracion
// La configuración es necesaria para configurar el scheduler
func NewCluster(name string, config configuration.Cluster) (*Cluster, error) {
	if config.Disabled {
		return nil, &ClusterDisabled{Name: name}
	}

	clusterScheduler, err := factory.Create(config.Scheduler.Type(), config.Scheduler.Parameters())
	if err != nil {
		util.Log.Fatalf("Error al crear el scheduler %s en %s. %s", config.Scheduler.Type(), name, err.Error())
	}

	c := &Cluster{
		name:      name,
		scheduler: clusterScheduler,
	}

	util.Log.WithFields(log.Fields{
		"cluster": name,
	}).Infof("Se creo un nuevo scheduler %s", config.Scheduler.Type())

	return c, nil
}

// GetScheduler retorna el scheduler que utiliza el cluster
func (c *Cluster) GetScheduler() scheduler.Scheduler {
	return c.scheduler
}
