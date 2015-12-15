package service

import "time"

// ServiceGroup agrupa un conjuntos de versiones de un servicios
type ServiceGroup struct {
	Id           string
	CreationDate time.Time
	Managers     map[string]*ServiceManager
}

// NewServiceGroup crea un nuevo contenedor de servicios con la fecha actual
func NewServiceGroup(id string) *ServiceGroup {
	container := &ServiceGroup{
		Id:           id,
		CreationDate: time.Now(),
		Managers:     make(map[string]*ServiceManager),
	}

	return container
}

// RegisterServiceManager registra una nuevo manejador de servicios
// Si el manager ya existia se retornara un error ServiceManagerAlreadyExist
func (s *ServiceGroup) RegisterServiceManager(clusterNames []string, params ServiceParameters) (*ServiceManager, error) {
	sm, err := NewServiceManager(clusterNames, params)
	if err != nil {
		return nil, err
	}

	if _, ok := s.Managers[sm.Id()]; ok {
		return nil, &ServiceManagerAlreadyExist{Service: s.Id, Version: params.Version}
	}

	sm.StartCheck()
	s.Managers[params.Version] = sm
	return sm, nil
}
