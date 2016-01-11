package service

import (
	"fmt"
	"regexp"
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
	updateInstancesMux sync.Mutex
	quitCheck          chan bool
	Version            string
	CreationDate       time.Time
	ImageName          string
	ImageTag           string
	interval           time.Duration
	instances          map[string]*Instance
	broadcaster        report.Broadcast
	threshold          int // limite de checks antes de marcar el servicio como fallido
	status             serviceStatus
	checkStatus        Checker
}

// NewServiceManager instancia un nuevo Manager
// Una vez creado un manager si la configuracion basica de este no permite crear una
// expresion regular valida para la busqueda de las imagenes asociadas (necesario para subscribirse en el updater)
// se retornara un error ImageNameRegexpError
// Si se pudo instancias bien el Manager se comenzará el chequeo de los servicios
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
		quitCheck:    make(chan bool),
		Version:      params.Version,
		CreationDate: time.Now(),
		ImageName:    params.ImageName,
		ImageTag:     params.ImageTag,
		interval:     interval,
		threshold:    threshold,
		instances:    make(map[string]*Instance),
		broadcaster:  broadcaster,
		status:       serviceStatus{},
	}

	_, err := sm.FullImageNameRegexp()
	if err != nil {
		return nil, &ImageNameRegexpError{Regexp: sm.FullImageName(), Message: err.Error()}
	}

	sm.checkStatus = sm.buildChecker(clusterNames, params)

	return sm, nil
}

func (s *Manager) buildChecker(clusterNames []string, params Parameters) Checker {
	minInstances := make(map[string]int)
	for _, clusterName := range clusterNames {
		minInstances[clusterName] = params.MinInstancesPerCluster[clusterName]
	}

	minInstancesChecker := &MinInstancesCheck{MinInstancesPerCluster: minInstances}
	atLeastXHost := &AtLeastXHostCheck{MinHosts: 2}
	minInstancesChecker.SetNext(atLeastXHost)
	return minInstancesChecker
}

// FullImageName obtiene el nombre completo de la imagen de este manager
// Si el el tag no existe se le concatenara como sufijo el tag por defecto "latest"
func (s *Manager) FullImageName() string {
	fullName := s.ImageName + ":"
	tag := "latest"
	if s.ImageTag != "" {
		tag = s.ImageTag
	}
	return fullName + tag
}

// FullImageNameRegexp construye y retorna la expresion regular necesaria para
// buscar todas las imagenes que pertenecen a este manager.
// Si la expresion regular generada es invalida se retornara un error
func (s *Manager) FullImageNameRegexp() (*regexp.Regexp, error) {
	fullName := s.FullImageName()
	return regexp.Compile("^" + fullName)
}

// ID retorna el identificador del manager el cual es:
// version:<nombre full de la imagen>
// Implementa ServiceUpdaterSubscriber
func (s *Manager) ID() string {
	return s.Version + ":" + s.FullImageName()
}

// Update implementa ServiceUpdaterSubscriber para recibir notificaciones
func (s *Manager) Update(data map[string]*monitor.ServiceUpdaterData) {
	s.updateInstancesMux.Lock()
	defer s.updateInstancesMux.Unlock()

	for k, v := range data {
		if v.InStatus(monitor.ServiceRemoved) {
			delete(s.instances, k)
		} else {
			instance := s.instances[k]
			if instance == nil {
				instance = &Instance{
					ID:           v.Origin().ID,
					CreationDate: time.Now(),
					// TODO FIX
					//Healthy:      v.Origin().Healthy(),
					ClusterID: v.ClusterID(),
					ImageName: v.Origin().ImageName,
					ImageTag:  v.Origin().ImageTag,
				}
			} else {
				// TODO FIX
				//instance.Healthy = v.Origin().Healthy()
			}
			s.instances[k] = instance
			logger.Instance().WithFields(log.Fields{
				"manager_id": s.ID(),
			}).Debugf("Servicio %s con data: %+v", v.LastAction(), instance)
		}
	}
}

// GetInstances retorna todas las instancias manejadas por este manager
func (s *Manager) GetInstances() map[string]*Instance {
	return s.instances
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
