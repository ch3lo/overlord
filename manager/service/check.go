package service

import "github.com/ch3lo/overlord/logger"

type Checker interface {
	id() string
	check(manager *Manager) bool
	Ok(manager *Manager) bool
	next() Checker
}

func checkHandler(c Checker, manager *Manager) bool {
	logger.Instance().Infoln("Handling Check", c.id())
	if c.check(manager) {
		if c.next() != nil {
			logger.Instance().Infoln("Checking", c.next().id())
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

func (c *MultiTagsChecker) Ok(manager *Manager) bool {
	return checkHandler(c, manager)
}

func (c *MultiTagsChecker) check(manager *Manager) bool {
	tags := make(map[string]bool)
	for _, v := range manager.instances {
		if v.Healthy {
			tags[v.ImageTag] = true
		}
	}

	logger.Instance().WithField("manager_id", manager.ID()).Debugf("Version %s Has multitags %t", manager.Version, len(tags) > 1)

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

func (c *MinInstancesCheck) Ok(manager *Manager) bool {
	return checkHandler(c, manager)
}

func (s *MinInstancesCheck) check(manager *Manager) bool {
	instancesPerCluster := make(map[string]int)
	for _, v := range manager.instances {
		if v.Healthy {
			instancesPerCluster[v.ClusterID]++
		}
	}

	for clusterId, minInstances := range s.MinInstancesPerCluster {
		if instancesPerCluster[clusterId] < minInstances {
			logger.Instance().WithField("manager_id", manager.ID()).Errorf("No hay un minimo de instancias para el cluster %s servicio %v", clusterId, manager)
			return false
		}
	}

	return true
}

type AtLeastXHostCheck struct {
	nextChecker Checker
	MinHosts    int
}

func (s *AtLeastXHostCheck) id() string {
	return "unique-host"
}

func (c *AtLeastXHostCheck) SetNext(next Checker) {
	c.nextChecker = next
}

func (c *AtLeastXHostCheck) next() Checker {
	return c.nextChecker
}

func (c *AtLeastXHostCheck) Ok(manager *Manager) bool {
	return checkHandler(c, manager)
}

func (s *AtLeastXHostCheck) check(manager *Manager) bool {
	hostsPerCluster := make(map[string]map[string]int)
	for _, v := range manager.instances {
		if v.Healthy {
			if _, ok := hostsPerCluster[v.ClusterID]; !ok {
				hostsPerCluster[v.ClusterID] = make(map[string]int)
			}
			hostsPerCluster[v.ClusterID][v.Host]++
		}
	}

	for k, v := range hostsPerCluster {
		if len(v) < s.MinHosts {
			logger.Instance().WithField("manager_id", manager.ID()).Errorf("No hay un minimo de servidores en el cluster %s ejecutando el servicio %v", k, manager)
			return false
		}
	}
	return true
}
