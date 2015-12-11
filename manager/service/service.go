package service

import (
	"regexp"
	"time"

	"github.com/ch3lo/overlord/monitor"
	"github.com/ch3lo/overlord/scheduler"
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
	Id        string
	Address   string
	Port      int
	Status    scheduler.ServiceInformationStatus // TODO No se debe depender de scheduler
	ClusterId string
	ImageName string
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

func (s *ServiceVersion) fullImageName() string {
	fullName := s.ImageName
	if s.ImageTag != "" {
		fullName += ":" + s.ImageTag
	}

	return fullName
}

func (s *ServiceVersion) fullImageNameRegexp() (*regexp.Regexp, error) {
	fullName := s.fullImageName() + ".*"
	return regexp.Compile(fullName)
}

func (s *ServiceVersion) Id() string {
	return s.Version + ":" + s.ImageName + ":" + s.ImageTag
}

func (s *ServiceVersion) Update(data map[string]*monitor.ServiceUpdaterData) {

	for k, v := range data {
		instance := &ServiceInstance{}
		instance = s.instances[k]
		if instance == nil {
			instance = &ServiceInstance{
				Id:        v.GetOrigin().Id,
				Status:    v.GetOrigin().Status,
				ClusterId: v.GetClusterId(),
				ImageName: v.GetOrigin().Image,
			}
		} else {
			instance.Status = v.GetOrigin().Status
		}
		s.instances[k] = instance

		util.Log.Printf("Servicio %s con data: %+v", v.GetLastAction(), instance)
	}
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

// AddServiceVersion registra una nueva version de servicio en el contenedor
// Si la version ya existia se retornara un error ServiceVersionAlreadyExist
func (s *ServiceContainer) AddServiceVersion(params ServiceParameters) (*ServiceVersion, error) {
	sv := &ServiceVersion{
		Version:                params.Version,
		CreationDate:           time.Now(),
		ImageName:              params.ImageName,
		ImageTag:               params.ImageTag,
		instances:              make(map[string]*ServiceInstance),
		MinInstancesPerCluster: params.MinInstancesPerCluster,
	}

	for key, _ := range s.Container {
		if key == sv.Id() {
			return nil, &ServiceVersionAlreadyExist{Service: s.Id, Version: params.Version}
		}
	}

	_, err := sv.fullImageNameRegexp()
	if err != nil {
		return nil, &ImageNameRegexpError{Regexp: sv.fullImageName(), Message: err.Error()}
	}

	s.Container[params.Version] = sv
	return sv, nil
}
