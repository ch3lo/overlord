package configuration

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Endpoint representa la configuración de un endpoint para notificaciones webhook
type Endpoint struct {
	Name      string        `yaml:"name"`      // nombre del endpoint
	Disabled  bool          `yaml:"disabled"`  // desabilita el endpoint
	URL       string        `yaml:"url"`       // URL del endpoint.
	Headers   http.Header   `yaml:"headers"`   // Headers que se deben agregar en cada request
	Timeout   time.Duration `yaml:"timeout"`   // HTTP timeout
	Threshold int           `yaml:"threshold"` // Cantidad de intentos antes de declarar un fallo
	Backoff   time.Duration `yaml:"backoff"`   // Tiempo de espera antes de volver a intentar el request
}

// Notifications agrupa un conjunto de endpoints http
type Notifications struct {
	Endpoints []Endpoint `yaml:"endpoints,omitempty"`
}

type Cluster struct {
	Disabled  bool      `yaml:"disabled"`
	Scheduler Scheduler `yaml:"scheduler"`
}

type Updater struct {
	Interval time.Duration `yaml:"interval,omitempty"`
}

type Configuration struct {
	Updater       *Updater              `yaml:"updater,omitempty"`
	Clusters      map[string]Cluster    `yaml:"cluster"`
	Notifications map[string]Parameters `yaml:"notifications,omitempty"`
}

type Scheduler map[string]Parameters

type Parameters map[string]interface{}

func (scheduler *Scheduler) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var schedulerMap map[string]Parameters
	err := unmarshal(&schedulerMap)
	if err == nil {
		if len(schedulerMap) > 1 {
			types := make([]string, 0, len(schedulerMap))
			for k := range schedulerMap {
				types = append(types, k)
			}

			if len(types) > 1 {
				return fmt.Errorf("Se debe configurar sólo un Scheduler. Schedulers: %v", types)
			}
		}
		*scheduler = schedulerMap
		return nil
	}

	var schedulerType string
	err = unmarshal(&schedulerType)
	if err == nil {
		*scheduler = Scheduler{schedulerType: Parameters{}}
		return nil
	}

	return err
}

func (scheduler Scheduler) MarshalYAML() (interface{}, error) {
	if scheduler.Parameters() == nil {
		return scheduler.Type(), nil
	}
	return map[string]Parameters(scheduler), nil
}

func (scheduler Scheduler) Parameters() Parameters {
	return scheduler[scheduler.Type()]
}

func (scheduler Scheduler) Type() string {
	var schedulerType []string

	// Return only key in this map
	for k := range scheduler {
		schedulerType = append(schedulerType, k)
	}
	if len(schedulerType) > 1 {
		panic("multiple storage drivers specified in configuration or environment: " + strings.Join(schedulerType, ", "))
	}
	if len(schedulerType) == 1 {
		return schedulerType[0]
	}
	return ""
}
