package cluster

import (
	"testing"

	"github.com/ch3lo/overlord/configuration"
	_ "github.com/ch3lo/overlord/scheduler/marathon"
	_ "github.com/ch3lo/overlord/scheduler/swarm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestCluster(t *testing.T) {
	suite.Run(t, new(ClusterSuite))
}

type ClusterSuite struct {
	suite.Suite
}

func (suite *ClusterSuite) SetupTest() {

}

func (suite *ClusterSuite) TestDefaultWithoutScheduler() {
	assert := assert.New(suite.T())
	config := configuration.Cluster{}
	c, err := NewCluster("asd", config)
	assert.Error(err)
	assert.Nil(c)
}

func (suite *ClusterSuite) TestDefaultWithScheduler() {
	assert := assert.New(suite.T())
	config := configuration.Cluster{
		Scheduler: configuration.Scheduler{
			"swarm": configuration.Parameters{
				"address": "3.3.3.3:8081",
			},
		},
	}
	c, err := NewCluster("asd", config)
	assert.Nil(err)
	assert.NotNil(c.GetScheduler())
	assert.Equal("asd", c.Id())
}

func (suite *ClusterSuite) TestClusterDisabled() {
	assert := assert.New(suite.T())
	config := configuration.Cluster{}
	config.Disabled = true
	c, err := NewCluster("asd", config)
	assert.IsType(new(ClusterDisabled), err)
	assert.Contains(err.Error(), "asd")
	assert.Nil(c)
}
