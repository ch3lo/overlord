package monitor

import (
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/ch3lo/overlord/scheduler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestCriteria(t *testing.T) {
	suite.Run(t, new(CriteriaSuite))
}

type CriteriaSuite struct {
	suite.Suite
	data map[string]*ServiceUpdaterData
}

func (suite *CriteriaSuite) SetupTest() {
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

func (suite *CriteriaSuite) assertLenCriteria(criteria ServiceChangeCriteria, length int) {
	result := criteria.MeetCriteria(suite.data)
	assert.Len(suite.T(), result, length)
}

func (suite *CriteriaSuite) assertImageNameAndImageTagRegexpCriteria(name string, length int) {
	suite.assertLenCriteria(&ImageNameAndImageTagRegexpCriteria{regexp.MustCompile(name)}, length)
}

func (suite *CriteriaSuite) TestImageNameAndImageTagRegexpCriteria() {
	suite.assertImageNameAndImageTagRegexpCriteria("registry.com/nombre_imagen", 1)
	suite.assertImageNameAndImageTagRegexpCriteria("nombre", 2)
	suite.assertImageNameAndImageTagRegexpCriteria("registry.com", 2)
	suite.assertImageNameAndImageTagRegexpCriteria("bad.registry", 0)
	suite.assertImageNameAndImageTagRegexpCriteria("registry.com/nombre_imagen:123", 0)
	suite.assertImageNameAndImageTagRegexpCriteria("registry.com/nombre_imagen:tag-123", 1)
	suite.assertImageNameAndImageTagRegexpCriteria("tag-123", 1)
	suite.assertImageNameAndImageTagRegexpCriteria(":tag-123$", 1)
	suite.assertImageNameAndImageTagRegexpCriteria(":tag$", 0)
}

func (suite *CriteriaSuite) TestStatusCriteria() {
	suite.assertLenCriteria(&StatusCriteria{ServiceAdded}, 1)
	suite.assertLenCriteria(&StatusCriteria{ServiceUpdated}, 1)
	suite.assertLenCriteria(&StatusCriteria{ServiceRemoved}, 0)
}

func (suite *CriteriaSuite) TestHealthyCriteria() {
	suite.assertLenCriteria(&HealthyCriteria{true}, 1)
	suite.assertLenCriteria(&HealthyCriteria{false}, 1)
}

func (suite *CriteriaSuite) TestAndCriteria() {
	suite.assertLenCriteria(&AndCriteria{
		&ImageNameAndImageTagRegexpCriteria{regexp.MustCompile("nombre")},
		&ImageNameAndImageTagRegexpCriteria{regexp.MustCompile("tag")},
	}, 2)
	suite.assertLenCriteria(&AndCriteria{&StatusCriteria{ServiceAdded}, &HealthyCriteria{true}}, 1)
	suite.assertLenCriteria(&AndCriteria{&StatusCriteria{ServiceAdded}, &HealthyCriteria{false}}, 0)
}

func (suite *CriteriaSuite) TestOrCriteria() {
	suite.assertLenCriteria(&OrCriteria{
		&ImageNameAndImageTagRegexpCriteria{regexp.MustCompile("nombre")},
		&ImageNameAndImageTagRegexpCriteria{regexp.MustCompile("tag")},
	}, 2)
	suite.assertLenCriteria(&OrCriteria{&HealthyCriteria{false}, &HealthyCriteria{true}}, 2)
	suite.assertLenCriteria(&OrCriteria{&StatusCriteria{ServiceAdded}, &HealthyCriteria{true}}, 1)
	suite.assertLenCriteria(&OrCriteria{&StatusCriteria{ServiceAdded}, &HealthyCriteria{false}}, 2)
}
