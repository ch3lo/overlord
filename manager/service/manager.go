package service

import (
	"regexp"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/manager/report"
	"github.com/ch3lo/overlord/monitor"
	"github.com/ch3lo/overlord/util"
)

type ServiceStatus struct {
	success          int
	failed           int
	consecutiveFails int
}

// ServiceManager es una estructura que contiene la información de una
// version de un servicio.
type ServiceManager struct {
	updateInstancesMux     sync.Mutex
	quitCheck              chan bool
	Version                string
	CreationDate           time.Time
	ImageName              string
	ImageTag               string
	interval               time.Duration
	instances              map[string]*ServiceInstance
	MinInstancesPerCluster map[string]int
	broadcaster            report.Broadcast
	threshold              int // limite de checks antes de marcar el servicio como fallido
	status                 ServiceStatus
}

func NewServiceManager(clusterNames []string, checkConfig configuration.Check, broadcaster report.Broadcast, params ServiceParameters) (*ServiceManager, error) {
	interval := time.Second * 10
	if checkConfig.Interval != 0 {
		interval = checkConfig.Interval
	}

	threshold := 5
	if checkConfig.Threshold != 0 {
		threshold = checkConfig.Threshold
	}

	sm := &ServiceManager{
		quitCheck:              make(chan bool),
		Version:                params.Version,
		CreationDate:           time.Now(),
		ImageName:              params.ImageName,
		ImageTag:               params.ImageTag,
		interval:               interval,
		threshold:              threshold,
		instances:              make(map[string]*ServiceInstance),
		MinInstancesPerCluster: make(map[string]int),
		broadcaster:            broadcaster,
		status:                 ServiceStatus{},
	}

	_, err := sm.FullImageNameRegexp()
	if err != nil {
		return nil, &ImageNameRegexpError{Regexp: sm.FullImageName(), Message: err.Error()}
	}

	for _, clusterName := range clusterNames {
		sm.MinInstancesPerCluster[clusterName] = params.MinInstancesPerCluster[clusterName]
	}

	return sm, nil
}

func (s *ServiceManager) FullImageName() string {
	fullName := s.ImageName
	if s.ImageTag != "" {
		fullName += ":" + s.ImageTag
	} else {
		fullName += ":latest"
	}

	return fullName
}

func (s *ServiceManager) FullImageNameRegexp() (*regexp.Regexp, error) {
	fullName := s.FullImageName()
	return regexp.Compile("^" + fullName)
}

func (s *ServiceManager) Id() string {
	return s.Version + ":" + s.FullImageName()
}

func (s *ServiceManager) Update(data map[string]*monitor.ServiceUpdaterData) {
	s.updateInstancesMux.Lock()
	defer s.updateInstancesMux.Unlock()

	for k, v := range data {
		if v.InStatus(monitor.SERVICE_REMOVED) {
			delete(s.instances, k)
		} else {
			instance := &ServiceInstance{}
			instance = s.instances[k]
			if instance == nil {
				instance = &ServiceInstance{
					Id:           v.Origin().Id,
					CreationDate: time.Now(),
					Healthy:      v.Origin().Healthy(),
					ClusterId:    v.ClusterId(),
					ImageName:    v.Origin().ImageName,
					ImageTag:     v.Origin().ImageTag,
				}
			} else {
				instance.Healthy = v.Origin().Healthy()
			}
			s.instances[k] = instance
			util.Log.WithFields(log.Fields{
				"manager_id": s.Id(),
			}).Debugf("Servicio %s con data: %+v", v.LastAction(), instance)
		}

	}
}

func (s *ServiceManager) GetInstances() map[string]*ServiceInstance {
	return s.instances
}

func (s *ServiceManager) StartCheck() {
	util.Log.WithField("manager_id", s.Id()).Infoln("Comenzando check")
	go s.checkInstances()
}

func (s *ServiceManager) StopCheck() {
	util.Log.WithField("manager_id", s.Id()).Infoln("Deteniendo check")
	s.quitCheck <- true
	<-s.quitCheck
	util.Log.WithField("manager_id", s.Id()).Infoln("Check detenido")
}

func (s *ServiceManager) check() {
	hasError := false
	if !s.checkMinInstances() {
		hasError = true
	}

	if hasError {
		s.status.consecutiveFails++
		s.status.failed++
	} else {
		s.status.consecutiveFails = 0
		s.status.success++
	}

	util.Log.WithField("manager_id", s.Id()).Debugf("Status del chequeo %+v - threshold %d", s.status, s.threshold)

	if s.threshold == s.status.consecutiveFails {
		s.broadcaster.Broadcast()
	}
}

func (s *ServiceManager) checkInstances() {
	for {
		select {
		case <-s.quitCheck:
			util.Log.WithField("manager_id", s.Id()).Infoln("Finalizando check")
			s.quitCheck <- true
			return
		case <-time.After(s.interval):
			s.check()
		}
	}
}

func (s *ServiceManager) checkMinInstances() bool {
	instancesPerCluster := make(map[string]int)
	for _, v := range s.instances {
		if v.Healthy {
			instancesPerCluster[v.ClusterId]++
		}
	}

	for clusterId, minInstances := range s.MinInstancesPerCluster {
		if instancesPerCluster[clusterId] < minInstances {
			util.Log.WithField("manager_id", s.Id()).Errorf("No hay un minimo de instancias para el cluster %s servicio %v", clusterId, s)
			return false
		}
	}

	return true
}

func (s *ServiceManager) hasMultiTags() bool {
	tags := make(map[string]bool)
	for _, v := range s.instances {
		tags[v.ImageTag] = true
	}

	util.Log.WithField("manager_id", s.Id()).Debugf("Version %s Has multitags %t", s.Version, len(tags) > 1)

	return len(tags) > 1
}
