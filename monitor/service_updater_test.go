package monitor

import (
	"fmt"
	"testing"

	"github.com/ch3lo/overlord/scheduler"

	"gopkg.in/check.v1"
)

func TestServiceUpdater(t *testing.T) { check.TestingT(t) }

type ServiceDataStatusSuite struct {
}

var _ = check.Suite(&ServiceDataStatusSuite{})

func (suite *ServiceDataStatusSuite) TestAreDistinct(c *check.C) {
	c.Assert(fmt.Sprint(ServiceUpdated), check.Equals, "ServiceUpdated")
	c.Assert(fmt.Sprint(ServiceAdded), check.Equals, "ServiceAdded")
	c.Assert(fmt.Sprint(ServiceRemoved), check.Equals, "ServiceRemoved")
	c.Assert(fmt.Sprint(ServiceUpdating), check.Equals, "ServiceUpdating")
}

type ServiceUpdaterDataSuite struct {
}

var _ = check.Suite(&ServiceUpdaterDataSuite{})

func (suite *ServiceUpdaterDataSuite) TestNew(c *check.C) {
	updater := NewServiceUpdaterData()
	c.Assert(updater.RegisterDate(), check.NotNil)
	c.Assert(updater.LastAction(), check.Equals, ServiceAdded)
	c.Assert(updater.ClusterID(), check.Equals, "")
	c.Assert(updater.Origin(), check.Equals, scheduler.ServiceInformation{})
	c.Assert(updater.InStatus(ServiceAdded), check.Equals, true)
}
