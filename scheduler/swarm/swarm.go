package swarm

// basado en https://github.com/docker/distribution/blob/603ffd58e18a9744679f741f2672dd9aea6babe0/registry/storage/driver/rados/rados.go

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/ch3lo/overlord/scheduler"
	"github.com/ch3lo/overlord/scheduler/factory"
	"github.com/ch3lo/overlord/util"
	"github.com/fsouza/go-dockerclient"
)

const schedulerId = "swarm"

func init() {
	factory.Register(schedulerId, &swarmCreator{})
}

// swarmCreator implementa la interfaz factory.SchedulerFactory
type swarmCreator struct{}

func (factory *swarmCreator) Create(parameters map[string]interface{}) (scheduler.Scheduler, error) {
	return NewFromParameters(parameters)
}

// SwarmParameters encapsula los parametros de configuracion de Swarm
type SwarmParameters struct {
	address   string
	tlsverify bool
	tlscacert string
	tlscert   string
	tlskey    string
}

// NewFromParameters construye un SwarmScheduler a partir de un mapeo de parámetros
// Al menos se debe pasar como parametro address
// Si se pasa tlsverify como true los parametros tlscacert, tlscert y tlskey también deben existir
func NewFromParameters(parameters map[string]interface{}) (*SwarmScheduler, error) {

	address, ok := parameters["address"]
	if !ok || fmt.Sprint(address) == "" {
		return nil, errors.New("Parametro address no existe")
	}

	tlsverify := false
	if tlsv, ok := parameters["tlsverify"]; ok {
		tlsverify, ok = tlsv.(bool)
		if !ok {
			return nil, fmt.Errorf("El parametro tlsverify debe ser un boolean")
		}
	}

	var tlscacert interface{}
	var tlscert interface{}
	var tlskey interface{}

	if tlsverify {
		tlscacert, ok = parameters["tlscacert"]
		if !ok || fmt.Sprint(tlscacert) == "" {
			return nil, errors.New("Parametro tlscacert no existe")
		}

		tlscert, ok = parameters["tlscert"]
		if !ok || fmt.Sprint(tlscert) == "" {
			return nil, errors.New("Parametro tlscert no existe")
		}

		tlskey, ok = parameters["tlskey"]
		if !ok || fmt.Sprint(tlskey) == "" {
			return nil, errors.New("Parametro tlskey no existe")
		}
	}

	params := SwarmParameters{
		address:   fmt.Sprint(address),
		tlsverify: tlsverify,
		tlscacert: fmt.Sprint(tlscacert),
		tlscert:   fmt.Sprint(tlscert),
		tlskey:    fmt.Sprint(tlskey),
	}

	return New(params)
}

// New construye un nuevo SwarmScheduler
func New(params SwarmParameters) (*SwarmScheduler, error) {

	swarm := new(SwarmScheduler)
	var err error
	util.Log.Debugf("Configuring Swarm with %+v", params)
	if params.tlsverify {
		swarm.client, err = docker.NewTLSClient(params.address, params.tlscert, params.tlskey, params.tlscacert)
	} else {
		swarm.client, err = docker.NewClient(params.address)
	}
	if err != nil {
		return nil, err
	}

	return swarm, nil
}

// SwarmScheduler es una implementacion de scheduler.Scheduler
// Permite el la comunicación con la API de Swarm
type SwarmScheduler struct {
	client *docker.Client
	tmp    string
}

func (ss *SwarmScheduler) Id() string {
	return schedulerId
}

func (ss *SwarmScheduler) IsAlive(id string) (bool, error) {
	container, err := ss.client.InspectContainer(id)
	if err != nil {
		return false, err
	}
	return container.State.Running && !container.State.Paused, nil
}

func (ss *SwarmScheduler) GetInstances(filter scheduler.FilterInstances) ([]scheduler.ServiceInformation, error) {
	// TODO implementar el uso del filtro
	containers, err := ss.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return nil, err
	}

	upRegexp := regexp.MustCompile("^[u|U]p")
	imageAndTagRegexp := regexp.MustCompile("^([\\w./_-]+)(?::([\\w._-]+))?$")

	var instances []scheduler.ServiceInformation
	for _, v := range containers {
		util.Log.WithField("scheduler", schedulerId).Debugf("Procesing container %+v", v)

		result := imageAndTagRegexp.FindStringSubmatch(v.Image)
		imageName := result[1]
		imageTag := "latest"
		if result[1] != "" {
			imageTag = result[2]
		}

		status := scheduler.SERVICE_DOWN
		if upRegexp.MatchString(v.Status) {
			status = scheduler.SERVICE_UP
		}

		instances = append(instances, scheduler.ServiceInformation{
			Id:        v.ID,
			Status:    status,
			ImageName: imageName,
			ImageTag:  imageTag,
		})
	}

	return instances, nil
}
