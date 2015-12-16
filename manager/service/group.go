package service

import (
	"time"

	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/manager/report"
)

// Group agrupa un conjuntos de versiones de un servicios
type Group struct {
	ID           string
	CreationDate time.Time
	Managers     map[string]*Manager
}

// NewServiceGroup crea un nuevo contenedor de servicios con la fecha actual
func NewServiceGroup(id string) *Group {
	container := &Group{
		ID:           id,
		CreationDate: time.Now(),
		Managers:     make(map[string]*Manager),
	}

	return container
}

// RegisterServiceManager registra una nuevo manejador de servicios
// Si el manager ya existia se retornara un error ServiceManagerAlreadyExist
func (s *Group) RegisterServiceManager(clusterNames []string, checkConfig configuration.Check, broadcaster report.Broadcast, params Parameters) (*Manager, error) {
	sm, err := NewServiceManager(clusterNames, checkConfig, broadcaster, params)
	if err != nil {
		return nil, err
	}

	if _, ok := s.Managers[sm.ID()]; ok {
		return nil, &ManagerAlreadyExist{Service: s.ID, Version: params.Version}
	}

	sm.StartCheck()
	s.Managers[params.Version] = sm
	return sm, nil
}
