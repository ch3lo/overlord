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

type serviceStatus struct {
	success          int
	failed           int
	consecutiveFails int
}

// ServiceManager es una estructura que contiene la información de una
// version de un servicio.
type ServiceManager struct {
	updateInstancesMux sync.Mutex
	quitCheck          chan bool
	Version            string
	CreationDate       time.Time
	ImageName          string
	ImageTag           string
	interval           time.Duration
	instances          map[string]*ServiceInstance
	broadcaster        report.Broadcast
	threshold          int // limite de checks antes de marcar el servicio como fallido
	status             serviceStatus
	checkStatus        Checker
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
		quitCheck:    make(chan bool),
		Version:      params.Version,
		CreationDate: time.Now(),
		ImageName:    params.ImageName,
		ImageTag:     params.ImageTag,
		interval:     interval,
		threshold:    threshold,
		instances:    make(map[string]*ServiceInstance),
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

func (s *ServiceManager) buildChecker(clusterNames []string, params ServiceParameters) Checker {
	minInstances := make(map[string]int)
	for _, clusterName := range clusterNames {
		minInstances[clusterName] = params.MinInstancesPerCluster[clusterName]
	}

	minInstancesChecker := &MinInstancesCheck{MinInstancesPerCluster: minInstances}
	return minInstancesChecker
}

func (s *ServiceManager) FullImageName() string {
	fullName := s.ImageName + ":"
	tag := "latest"
	if s.ImageTag != "" {
		tag = s.ImageTag
	}
	return fullName + tag
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
			instance := s.instances[k]
			if instance == nil {
				instance = &ServiceInstance{
					Id:           v.Origin().ID,
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
	if s.checkStatus.Ok(s) {
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
