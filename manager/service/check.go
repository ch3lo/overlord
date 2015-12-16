package service

import (
	"github.com/ch3lo/overlord/util"
)

type Checker interface {
	id() string
	check(manager *ServiceManager) bool
	Ok(manager *ServiceManager) bool
	next() Checker
}

func checkHandler(c Checker, manager *ServiceManager) bool {
	util.Log.Infoln("Checking", c.id())
	if c.check(manager) {
		if c.next() != nil {
			return c.next().check(manager)
		}
		return true
	}
	return false
}

type MultiTagsChecker struct {
	nextChecker Checker
}

func (s *MultiTagsChecker) id() string {
	return "multi-tags"
}

func (c *MultiTagsChecker) SetNext(next Checker) {
	c.nextChecker = next
}

func (c *MultiTagsChecker) next() Checker {
	return c.nextChecker
}

func (c *MultiTagsChecker) Ok(manager *ServiceManager) bool {
	return checkHandler(c, manager)
}

func (c *MultiTagsChecker) check(manager *ServiceManager) bool {
	tags := make(map[string]bool)
	for _, v := range manager.instances {
		tags[v.ImageTag] = true
	}

	util.Log.WithField("manager_id", manager.Id()).Debugf("Version %s Has multitags %t", manager.Version, len(tags) > 1)

	return len(tags) > 1
}

type MinInstancesCheck struct {
	nextChecker            Checker
	MinInstancesPerCluster map[string]int
}

func (s *MinInstancesCheck) id() string {
	return "min-instances"
}

func (c *MinInstancesCheck) SetNext(next Checker) {
	c.nextChecker = next
}

func (c *MinInstancesCheck) next() Checker {
	return c.nextChecker
}

func (c *MinInstancesCheck) Ok(manager *ServiceManager) bool {
	return checkHandler(c, manager)
}

func (s *MinInstancesCheck) check(manager *ServiceManager) bool {
	instancesPerCluster := make(map[string]int)
	for _, v := range manager.instances {
		if v.Healthy {
			instancesPerCluster[v.ClusterId]++
		}
	}

	for clusterId, minInstances := range s.MinInstancesPerCluster {
		if instancesPerCluster[clusterId] < minInstances {
			util.Log.WithField("manager_id", manager.Id()).Errorf("No hay un minimo de instancias para el cluster %s servicio %v", clusterId, manager)
			return false
		}
	}

	return true
}
