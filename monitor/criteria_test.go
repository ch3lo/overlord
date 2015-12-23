package monitor

import (
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/ch3lo/overlord/scheduler"

	"gopkg.in/check.v1"
)

func TestCriteria(t *testing.T) { check.TestingT(t) }

type CriteriaSuite struct {
	data map[string]*ServiceUpdaterData
}

var _ = check.Suite(&CriteriaSuite{})

func (suite *CriteriaSuite) SetUpTest(c *check.C) {
	os.Clearenv()
	suite.data = make(map[string]*ServiceUpdaterData)

	date1, _ := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	suite.data["qwerty12345"] = &ServiceUpdaterData{
		registerDate: date1,
		lastUpdate:   date1,
		lastAction:   ServiceAdded,
		clusterID:    "wdc",
		origin: scheduler.ServiceInformation{
			ID:            "qwerty12345",
			ImageName:     "registry.com/nombre_imagen",
			ImageTag:      "tag-123",
			Host:          "thor1",
			ContainerName: "container_name1",
			Status:        scheduler.ServiceUp,
		},
	}

	date2, _ := time.Parse(time.RFC3339, "2013-11-01T22:08:41+00:00")
	suite.data["asdasd"] = &ServiceUpdaterData{
		registerDate: date2,
		lastUpdate:   date2,
		lastAction:   ServiceUpdated,
		clusterID:    "dal",
		origin: scheduler.ServiceInformation{
			ID:            "asdasd",
			ImageName:     "registry.com/imagen_nombre",
			ImageTag:      "tag-234",
			Host:          "thor2",
			ContainerName: "container_name2",
			Status:        scheduler.ServiceDown,
		},
	}
}

func (suite *CriteriaSuite) assertLenCriteria(c *check.C, criteria ServiceChangeCriteria, length int) {
	result := criteria.MeetCriteria(suite.data)
	c.Assert(result, check.HasLen, length)
}

func (suite *CriteriaSuite) assertImageNameAndImageTagRegexpCriteria(c *check.C, name string, length int) {
	suite.assertLenCriteria(c, &ImageNameAndImageTagRegexpCriteria{regexp.MustCompile(name)}, length)
}

func (suite *CriteriaSuite) TestImageNameAndImageTagRegexpCriteria(c *check.C) {
	suite.assertImageNameAndImageTagRegexpCriteria(c, "registry.com/nombre_imagen", 1)
	suite.assertImageNameAndImageTagRegexpCriteria(c, "nombre", 2)
	suite.assertImageNameAndImageTagRegexpCriteria(c, "registry.com", 2)
	suite.assertImageNameAndImageTagRegexpCriteria(c, "bad.registry", 0)
	suite.assertImageNameAndImageTagRegexpCriteria(c, "registry.com/nombre_imagen:123", 0)
	suite.assertImageNameAndImageTagRegexpCriteria(c, "registry.com/nombre_imagen:tag-123", 1)
	suite.assertImageNameAndImageTagRegexpCriteria(c, "tag-123", 1)
	suite.assertImageNameAndImageTagRegexpCriteria(c, ":tag-123$", 1)
	suite.assertImageNameAndImageTagRegexpCriteria(c, ":tag$", 0)
}

func (suite *CriteriaSuite) TestStatusCriteria(c *check.C) {
	suite.assertLenCriteria(c, &StatusCriteria{ServiceAdded}, 1)
	suite.assertLenCriteria(c, &StatusCriteria{ServiceUpdated}, 1)
	suite.assertLenCriteria(c, &StatusCriteria{ServiceRemoved}, 0)
}

func (suite *CriteriaSuite) TestHealthyCriteria(c *check.C) {
	suite.assertLenCriteria(c, &HealthyCriteria{true}, 1)
	suite.assertLenCriteria(c, &HealthyCriteria{false}, 1)
}

func (suite *CriteriaSuite) TestAndCriteria(c *check.C) {
	suite.assertLenCriteria(c, &AndCriteria{
		&ImageNameAndImageTagRegexpCriteria{regexp.MustCompile("nombre")},
		&ImageNameAndImageTagRegexpCriteria{regexp.MustCompile("tag")},
	}, 2)
	suite.assertLenCriteria(c, &AndCriteria{&StatusCriteria{ServiceAdded}, &HealthyCriteria{true}}, 1)
	suite.assertLenCriteria(c, &AndCriteria{&StatusCriteria{ServiceAdded}, &HealthyCriteria{false}}, 0)
}

func (suite *CriteriaSuite) TestOrCriteria(c *check.C) {
	suite.assertLenCriteria(c, &OrCriteria{
		&ImageNameAndImageTagRegexpCriteria{regexp.MustCompile("nombre")},
		&ImageNameAndImageTagRegexpCriteria{regexp.MustCompile("tag")},
	}, 2)
	suite.assertLenCriteria(c, &OrCriteria{&HealthyCriteria{false}, &HealthyCriteria{true}}, 2)
	suite.assertLenCriteria(c, &OrCriteria{&StatusCriteria{ServiceAdded}, &HealthyCriteria{true}}, 1)
	suite.assertLenCriteria(c, &OrCriteria{&StatusCriteria{ServiceAdded}, &HealthyCriteria{false}}, 2)
}
