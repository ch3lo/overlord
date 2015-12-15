package service

import (
	"regexp"
	"sync"
	"time"

	"github.com/ch3lo/overlord/monitor"
	"github.com/ch3lo/overlord/util"
)

// ServiceManager es una estructura que contiene la información de una
// version de un servicio.
type ServiceManager struct {
	updateInstancesMux     sync.Mutex
	quitCheck              chan bool
	Version                string
	CreationDate           time.Time
	ImageName              string
	ImageTag               string
	instances              map[string]*ServiceInstance
	MinInstancesPerCluster map[string]int
	serviceUpdater         *monitor.ServiceUpdater
}

func NewServiceManager(clusterNames []string, params ServiceParameters) (*ServiceManager, error) {
	sm := &ServiceManager{
		quitCheck:              make(chan bool),
		Version:                params.Version,
		CreationDate:           time.Now(),
		ImageName:              params.ImageName,
		ImageTag:               params.ImageTag,
		instances:              make(map[string]*ServiceInstance),
		MinInstancesPerCluster: make(map[string]int),
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
		instance := &ServiceInstance{}
		instance = s.instances[k]
		if instance == nil {
			instance = &ServiceInstance{
				Id:           v.GetOrigin().Id,
				CreationDate: time.Now(),
				Status:       v.GetOrigin().Status,
				ClusterId:    v.GetClusterId(),
				ImageName:    v.GetOrigin().ImageName,
				ImageTag:     v.GetOrigin().ImageTag,
			}
		} else {
			instance.Status = v.GetOrigin().Status
		}
		s.instances[k] = instance
		util.Log.WithField("instance_id", instance.Id).Debugf("Servicio %s con data: %+v", v.GetLastAction(), instance)
	}
}

func (s *ServiceManager) GetInstances() map[string]*ServiceInstance {
	return s.instances
}

func (s *ServiceManager) StartCheck() {
	go s.checkInstances()
}

func (s *ServiceManager) StopCheck() {
	util.Log.Infoln("Deteniendo check")
	s.quitCheck <- true
	<-s.quitCheck
	util.Log.Infoln("Check detenido")
}

func (s *ServiceManager) checkInstances() {
	for {
		select {
		case <-s.quitCheck:
			util.Log.Infoln("Finalizando check")
			s.quitCheck <- true
			return
		case <-time.After(10 * time.Second):
			util.Log.Infoln("CHECKING INSTANCES")
			s.hasMultiTags()
			s.checkMinInstances()
		}
	}
}

func (s *ServiceManager) checkMinInstances() {
	instancesPerCluster := make(map[string]int)
	for _, v := range s.instances {
		instancesPerCluster[v.ClusterId]++
	}

	for clusterId, minInstances := range s.MinInstancesPerCluster {
		if instancesPerCluster[clusterId] < minInstances {
			util.Log.Errorf("No hay un minimo de instancias para el cluster %s servicio %v", clusterId, s)
		}
	}
}

func (s *ServiceManager) hasMultiTags() bool {
	tags := make(map[string]bool)
	for _, v := range s.instances {
		tags[v.ImageTag] = true
	}

	util.Log.WithField("version", s.Version).Debugf("Has multitags %t", len(tags) > 1)

	return len(tags) > 1
}
