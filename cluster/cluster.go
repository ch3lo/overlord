package cluster

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/logger"
	"github.com/latam-airlines/mesos-framework-factory"
	"github.com/latam-airlines/mesos-framework-factory/factory"
)

type Cluster struct {
	id        string
	scheduler framework.Framework
}

// NewCluster crea un nuevo cluster a partir de un id y parametros de configuracion
// La configuración es necesaria para configurar el scheduler
func NewCluster(custerId string, config configuration.Cluster) (*Cluster, error) {
	if config.Disabled {
		return nil, &ClusterDisabled{Name: custerId}
	}

	clusterScheduler, err := factory.Create(config.Scheduler.Type(), config.Scheduler.Parameters())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error al crear el scheduler %s en %s. %s", config.Scheduler.Type(), custerId, err.Error()))
	}

	c := &Cluster{
		id:        custerId,
		scheduler: clusterScheduler,
	}

	logger.Instance().WithFields(log.Fields{
		"cluster": custerId,
	}).Infof("Se creo un nuevo scheduler %s", config.Scheduler.Type())

	return c, nil
}

func (c *Cluster) Id() string {
	return c.id
}

// GetScheduler retorna el scheduler que utiliza el cluster
func (c *Cluster) GetScheduler() framework.Framework {
	return c.scheduler
}
