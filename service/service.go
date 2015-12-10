package service

import (
	"time"

	"github.com/ch3lo/overlord/updater"
	"github.com/ch3lo/overlord/util"
)

// ServiceParameters es una estructura que encapsula los parametros
// de configuración de un nuevo servicio
type ServiceParameters struct {
	Id                     string
	Version                string
	ImageName              string
	ImageTag               string
	MinInstancesPerCluster map[string]int
}

// ServiceInstance contiene la información de una instancia de un servicio
type ServiceInstance struct {
	Id          string
	Address     string
	Port        int
	Status      string
	ClusterName string
}

// ServiceVersion es una estructura que contiene la información de una
// version de un servicio.
type ServiceVersion struct {
	Version                string
	CreationDate           time.Time
	ImageName              string
	ImageTag               string
	instances              map[string]*ServiceInstance
	MinInstancesPerCluster map[string]int
}

func (s *ServiceVersion) Id() string {
	return s.Version + ":" + s.ImageName + ":" + s.ImageTag
}

func (s *ServiceVersion) Update(data map[string]*updater.ServiceUpdaterData) {
	util.Log.Println("Servicio actualizado", s.Id())
}

// ServiceContainer agrupa un conjuntos de versiones de un servicio bajo el parametro Container
type ServiceContainer struct {
	Id           string
	CreationDate time.Time
	Container    map[string]*ServiceVersion
}

// NewServiceContainer crea un nuevo contenedor de servicios con la fecha actual
func NewServiceContainer(id string) *ServiceContainer {
	container := &ServiceContainer{
		Id:           id,
		CreationDate: time.Now(),
		Container:    make(map[string]*ServiceVersion),
	}

	return container
}

// RegisterServiceVersion registra una nueva version de servicio en el contenedor
// Si la version ya existia se retornara un error ServiceVersionAlreadyExist
func (s *ServiceContainer) RegisterServiceVersion(params ServiceParameters) (*ServiceVersion, error) {
	for key, _ := range s.Container {
		if key == params.Version {
			return nil, &ServiceVersionAlreadyExist{Service: s.Id, Version: params.Version}
		}
	}

	sv := &ServiceVersion{
		Version:                params.Version,
		CreationDate:           time.Now(),
		ImageName:              params.ImageName,
		ImageTag:               params.ImageTag,
		instances:              make(map[string]*ServiceInstance),
		MinInstancesPerCluster: params.MinInstancesPerCluster,
	}

	s.Container[params.Version] = sv
	return sv, nil
}
