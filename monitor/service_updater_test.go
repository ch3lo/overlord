package monitor

import (
	"fmt"
	"testing"
	"time"

	"github.com/ch3lo/overlord/cluster"
	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestServiceUpdater(t *testing.T) {
	suite.Run(t, new(ServiceDataStatusSuite))
	suite.Run(t, new(ServiceUpdaterDataSuite))
	suite.Run(t, new(ServiceUpdaterSuite))
}

type ServiceDataStatusSuite struct {
	suite.Suite
}

func (suite *ServiceDataStatusSuite) TestAreDistinct() {
	assert := assert.New(suite.T())
	assert.Equal(fmt.Sprint(ServiceUpdated), "ServiceUpdated")
	assert.Equal(fmt.Sprint(ServiceAdded), "ServiceAdded")
	assert.Equal(fmt.Sprint(ServiceRemoved), "ServiceRemoved")
	assert.Equal(fmt.Sprint(ServiceUpdating), "ServiceUpdating")
}

type ServiceUpdaterDataSuite struct {
	suite.Suite
}

func (suite *ServiceUpdaterDataSuite) TestNew() {
	assert := assert.New(suite.T())
	updater := NewServiceUpdaterData()
	assert.NotNil(updater.RegisterDate())
	assert.Equal(updater.LastAction(), ServiceAdded)
	assert.Equal(updater.ClusterID(), "")
	assert.Equal(updater.Origin(), scheduler.ServiceInformation{})
	assert.Equal(updater.InStatus(ServiceAdded), true)
}

type ServiceUpdaterSuite struct {
	suite.Suite
}

func (suite *ServiceUpdaterSuite) TestNewServiceUpdater() {
	assert := assert.New(suite.T())
	config := configuration.Updater{}
	assert.Panics(func() { NewServiceUpdater(config, nil) })
	c := make(map[string]*cluster.Cluster)
	assert.Panics(func() { NewServiceUpdater(config, c) })

	c["asd"] = &cluster.Cluster{}
	updater := NewServiceUpdater(config, c)
	assert.NotNil(updater)
	assert.Equal(10*time.Second, updater.interval)
	assert.NotNil(updater.subscribers)
	assert.NotNil(updater.subscriberCriteria)
	assert.NotNil(updater.services)
	assert.Equal(c, updater.clusters)
}
