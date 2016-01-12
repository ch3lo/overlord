package service

import (
	"regexp"
	"time"

	"github.com/ch3lo/overlord/monitor"
)

type ConstraintsParams struct {
	ImageName              string
	MinInstancesPerCluster map[string]int
}

// Parameters es una estructura que encapsula los parametros
// de configuración de un nuevo servicio
type Parameters struct {
	ID          string
	Version     string
	Constraints ConstraintsParams
}

func (p *Parameters) BuildCriteria() (monitor.ServiceChangeCriteria, error) {
	var criteria monitor.ServiceChangeCriteria
	if p.Constraints.ImageName != "" {
		reg := "^" + p.Constraints.ImageName
		imageNameRegexp, err := regexp.Compile(reg)
		if err != nil {
			return nil, &ImageNameRegexpError{Regexp: reg, Message: err.Error()}
		}
		criteria = &monitor.ImageNameAndImageTagRegexpCriteria{imageNameRegexp}
	}
	return criteria, nil
}

// Instance contiene la información de una instancia de un servicio
type Instance struct {
	ID           string
	ImageName    string
	ImageTag     string
	CreationDate time.Time
	Host         string
	Healthy      bool
	ClusterID    string
}

// Application agrupa un conjuntos de versiones de un servicios
type AppMajor struct {
	ID           string
	Version      string
	CreationDate time.Time
	Instances    map[string]*Instance
	Constraints  ConstraintsParams
}

// NewServiceGroup crea un nuevo contenedor de servicios con la fecha actual
func NewAppMajor(params Parameters) *AppMajor {
	app := &AppMajor{
		ID:           params.ID,
		Version:      params.Version,
		CreationDate: time.Now(),
		Instances:    make(map[string]*Instance),
		Constraints:  params.Constraints,
	}

	return app
}
