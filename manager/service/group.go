package service

import "time"

// ServiceGroup agrupa un conjuntos de versiones de un servicios
type ServiceGroup struct {
	Id           string
	CreationDate time.Time
	Container    map[string]*ServiceManager
}

// NewServiceGroup crea un nuevo contenedor de servicios con la fecha actual
func NewServiceGroup(id string) *ServiceGroup {
	container := &ServiceGroup{
		Id:           id,
		CreationDate: time.Now(),
		Container:    make(map[string]*ServiceManager),
	}

	return container
}

// RegisterServiceManager registra una nuevo manejador de servicios
// Si el manager ya existia se retornara un error ServiceManagerAlreadyExist
func (s *ServiceGroup) RegisterServiceManager(params ServiceParameters) (*ServiceManager, error) {
	sv := &ServiceManager{
		Version:                params.Version,
		CreationDate:           time.Now(),
		ImageName:              params.ImageName,
		ImageTag:               params.ImageTag,
		instances:              make(map[string]*ServiceInstance),
		MinInstancesPerCluster: params.MinInstancesPerCluster,
	}

	for key, _ := range s.Container {
		if key == sv.Id() {
			return nil, &ServiceManagerAlreadyExist{Service: s.Id, Version: params.Version}
		}
	}

	_, err := sv.FullImageNameRegexp()
	if err != nil {
		return nil, &ImageNameRegexpError{Regexp: sv.FullImageName(), Message: err.Error()}
	}

	s.Container[params.Version] = sv
	return sv, nil
}
