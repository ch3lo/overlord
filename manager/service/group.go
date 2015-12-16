package service

import (
	"time"

	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/manager/report"
)

// ServiceGroup agrupa un conjuntos de versiones de un servicios
type ServiceGroup struct {
	ID           string
	CreationDate time.Time
	Managers     map[string]*ServiceManager
}

// NewServiceGroup crea un nuevo contenedor de servicios con la fecha actual
func NewServiceGroup(id string) *ServiceGroup {
	container := &ServiceGroup{
		ID:           id,
		CreationDate: time.Now(),
		Managers:     make(map[string]*ServiceManager),
	}

	return container
}

// RegisterServiceManager registra una nuevo manejador de servicios
// Si el manager ya existia se retornara un error ServiceManagerAlreadyExist
func (s *ServiceGroup) RegisterServiceManager(clusterNames []string, checkConfig configuration.Check, broadcaster report.Broadcast, params ServiceParameters) (*ServiceManager, error) {
	sm, err := NewServiceManager(clusterNames, checkConfig, broadcaster, params)
	if err != nil {
		return nil, err
	}

	if _, ok := s.Managers[sm.Id()]; ok {
		return nil, &ServiceManagerAlreadyExist{Service: s.ID, Version: params.Version}
	}

	sm.StartCheck()
	s.Managers[params.Version] = sm
	return sm, nil
}
