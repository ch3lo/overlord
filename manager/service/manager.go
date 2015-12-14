package service

import (
	"regexp"
	"time"

	"github.com/ch3lo/overlord/monitor"
	"github.com/ch3lo/overlord/util"
)

// ServiceManager es una estructura que contiene la información de una
// version de un servicio.
type ServiceManager struct {
	Version                string
	CreationDate           time.Time
	ImageName              string
	ImageTag               string
	instances              map[string]*ServiceInstance
	MinInstancesPerCluster map[string]int
	serviceUpdater         *monitor.ServiceUpdater
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

	util.Log.WithField("version", s.Version).Debugf("Has multitags %t", s.hasMultiTags())
}

func (s *ServiceManager) GetInstances() map[string]*ServiceInstance {
	return s.instances
}

func (s *ServiceManager) hasMultiTags() bool {
	tags := make(map[string]bool)
	for _, v := range s.instances {
		tags[v.ImageTag] = true
	}

	return len(tags) > 1
}
