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

const schedulerID = "swarm"

func init() {
	factory.Register(schedulerID, &swarmCreator{})
}

// swarmCreator implementa la interfaz factory.SchedulerFactory
type swarmCreator struct{}

func (factory *swarmCreator) Create(parameters map[string]interface{}) (scheduler.Scheduler, error) {
	return NewFromParameters(parameters)
}

// parameters encapsula los parametros de configuracion de Swarm
type parameters struct {
	address   string
	tlsverify bool
	tlscacert string
	tlscert   string
	tlskey    string
}

// NewFromParameters construye un Scheduler a partir de un mapeo de parámetros
// Al menos se debe pasar como parametro address, ya que si no existe se retornara un error
// Si se pasa tlsverify como true los parametros tlscacert, tlscert y tlskey también deben existir
func NewFromParameters(params map[string]interface{}) (*Scheduler, error) {

	address, ok := params["address"]
	if !ok || fmt.Sprint(address) == "" {
		return nil, errors.New("Parametro address no existe")
	}

	tlsverify := false
	if tlsv, ok := params["tlsverify"]; ok {
		tlsverify, ok = tlsv.(bool)
		if !ok {
			return nil, fmt.Errorf("El parametro tlsverify debe ser un boolean")
		}
	}

	var tlscacert interface{}
	var tlscert interface{}
	var tlskey interface{}

	if tlsverify {
		tlscacert, ok = params["tlscacert"]
		if !ok || fmt.Sprint(tlscacert) == "" {
			return nil, errors.New("Parametro tlscacert no existe")
		}

		tlscert, ok = params["tlscert"]
		if !ok || fmt.Sprint(tlscert) == "" {
			return nil, errors.New("Parametro tlscert no existe")
		}

		tlskey, ok = params["tlskey"]
		if !ok || fmt.Sprint(tlskey) == "" {
			return nil, errors.New("Parametro tlskey no existe")
		}
	}

	p := parameters{
		address:   fmt.Sprint(address),
		tlsverify: tlsverify,
		tlscacert: fmt.Sprint(tlscacert),
		tlscert:   fmt.Sprint(tlscert),
		tlskey:    fmt.Sprint(tlskey),
	}

	return New(p)
}

// New instancia un nuevo cliente de Swarm
func New(params parameters) (*Scheduler, error) {

	swarm := new(Scheduler)
	var err error
	util.Log.Debugf("Configurando Swarm con los parametros %+v", params)
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

// Scheduler es una implementacion de scheduler.Scheduler
// Permite el la comunicación con la API de Swarm
type Scheduler struct {
	client *docker.Client
	tmp    string
}

// ID retorna el identificador del scheduler Swarm
func (ss *Scheduler) ID() string {
	return schedulerID
}

// IsAlive retorna si el servicio se encuentra en estado OK
func (ss *Scheduler) IsAlive(id string) (bool, error) {
	container, err := ss.client.InspectContainer(id)
	if err != nil {
		return false, err
	}
	return container.State.Running && !container.State.Paused, nil
}

// Instances retorna todas las instancias que maneja el scheduler
// Parseando la informacion obtenida para cumplir con la interfaz scheduler.ServiceInformation
func (ss *Scheduler) Instances() ([]scheduler.ServiceInformation, error) {
	// TODO implementar el uso de filtro con criterios
	containers, err := ss.client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return nil, err
	}

	upRegexp := regexp.MustCompile("^[u|U]p")
	imageAndTagRegexp := regexp.MustCompile("^([\\w./_-]+)(?::([\\w._-]+))?$")
	hostAndContainerName := regexp.MustCompile("^(?:/([\\w|_-]+))?/([\\w|_-]+)$")

	var instances []scheduler.ServiceInformation
	for _, v := range containers {
		util.Log.WithField("scheduler", schedulerID).Debugf("Procesing container %+v", v)

		result := imageAndTagRegexp.FindStringSubmatch(v.Image)
		imageName := result[1]
		imageTag := "latest"
		if result[1] != "" {
			imageTag = result[2]
		}

		status := scheduler.ServiceDown
		if upRegexp.MatchString(v.Status) {
			status = scheduler.ServiceUp
		}

		result = hostAndContainerName.FindStringSubmatch(v.Names[0])
		host := "unknown"
		if result[1] != "" {
			host = result[1]
		}
		containerName := result[2]

		instances = append(instances, scheduler.ServiceInformation{
			ID:            v.ID,
			Status:        status,
			ImageName:     imageName,
			ImageTag:      imageTag,
			Host:          host,
			ContainerName: containerName,
		})
	}

	return instances, nil
}
