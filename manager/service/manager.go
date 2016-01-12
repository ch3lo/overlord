package service

import (
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/logger"
	"github.com/ch3lo/overlord/manager/report"
	"github.com/ch3lo/overlord/monitor"
)

type serviceStatus struct {
	success          int
	failed           int
	consecutiveFails int
}

// Manager es una estructura que contiene la información de una
// version de un servicio.
type Manager struct {
	id                 string
	updateInstancesMux sync.Mutex
	quitCheck          chan bool
	Version            string
	CreationDate       time.Time
	interval           time.Duration
	broadcaster        report.Broadcast
	threshold          int // limite de checks antes de marcar el servicio como fallido
	status             serviceStatus
	checkStatus        Checker
	App                *AppMajor
}

// NewServiceManager instancia un nuevo Manager y el chequeo de los servicios asociados
func NewServiceManager(clusterNames []string, checkConfig configuration.Check, broadcaster report.Broadcast, params Parameters) (*Manager, error) {
	interval := time.Second * 10
	if checkConfig.Interval != 0 {
		interval = checkConfig.Interval
	}

	threshold := 5
	if checkConfig.Threshold != 0 {
		threshold = checkConfig.Threshold
	}

	sm := &Manager{
		id:           params.ID + "#" + params.Version,
		quitCheck:    make(chan bool),
		Version:      params.Version,
		CreationDate: time.Now(),
		interval:     interval,
		threshold:    threshold,
		broadcaster:  broadcaster,
		status:       serviceStatus{},
		App:          NewAppMajor(params),
	}

	sm.checkStatus = sm.buildChecker(clusterNames, params)

	return sm, nil
}

func (s *Manager) buildChecker(clusterNames []string, params Parameters) Checker {
	minInstances := make(map[string]int)
	for _, clusterName := range clusterNames {
		minInstances[clusterName] = params.Constraints.MinInstancesPerCluster[clusterName]
	}

	minInstancesChecker := &MinInstancesCheck{MinInstancesPerCluster: minInstances}
	atLeastXHost := &AtLeastXHostCheck{MinHosts: 2}
	minInstancesChecker.SetNext(atLeastXHost)
	return minInstancesChecker
}

// ID retorna el identificador del manager el cual es:
// <id de la app>#<version mayor de la app>
// Implementa ServiceUpdaterSubscriber
func (s *Manager) ID() string {
	return s.id
}

// Update implementa ServiceUpdaterSubscriber para recibir notificaciones
func (s *Manager) Update(data map[string]*monitor.ServiceUpdaterData) {
	s.updateInstancesMux.Lock()
	defer s.updateInstancesMux.Unlock()

	for _, v := range data {
		for _, w := range v.Origin().Instances {
			if v.InStatus(monitor.ServiceRemoved) {
				delete(s.App.Instances, w.ID)
			} else {
				instance, ok := s.App.Instances[w.ID]
				if ok {
					instance.Healthy = w.Healthy()
				} else {
					instance = &Instance{
						ID:           w.ID,
						CreationDate: time.Now(),
						Healthy:      w.Healthy(),
						ClusterID:    v.ClusterID(),
						ImageName:    v.Origin().ImageName,
						ImageTag:     v.Origin().ImageTag,
					}

				}

				s.App.Instances[w.ID] = instance
				logger.Instance().WithFields(log.Fields{
					"manager_id": s.ID(),
				}).Debugf("Servicio %s con data: %+v", v.LastAction(), instance)
			}
		}
	}
}

// StartCheck comienza el chequeo de los servicios
func (s *Manager) StartCheck() {
	logger.Instance().WithField("manager_id", s.ID()).Infoln("Comenzando check")
	go s.checkInstances()
}

// StopCheck detiene el chequeo de los servicios
func (s *Manager) StopCheck() {
	logger.Instance().WithField("manager_id", s.ID()).Infoln("Deteniendo check")
	s.quitCheck <- true
	<-s.quitCheck
	logger.Instance().WithField("manager_id", s.ID()).Infoln("Check detenido")
}

func (s *Manager) check() {
	if s.checkStatus.Ok(s) {
		s.status.consecutiveFails = 0
		s.status.success++
	} else {
		s.status.consecutiveFails++
		s.status.failed++
	}

	logger.Instance().WithField("manager_id", s.ID()).Debugf("Status del chequeo %+v - threshold %d", s.status, s.threshold)

	if s.threshold == s.status.consecutiveFails {
		var query = []byte(fmt.Sprintf("Status del chequeo %+v - threshold %d", s.status, s.threshold))
		s.broadcaster.Broadcast(query)
	}
}

func (s *Manager) checkInstances() {
	for {
		select {
		case <-s.quitCheck:
			logger.Instance().WithField("manager_id", s.ID()).Infoln("Finalizando check")
			s.quitCheck <- true
			return
		case <-time.After(s.interval):
			s.check()
		}
	}
}
