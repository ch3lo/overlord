package service

import (
	"time"

	"github.com/ch3lo/overlord/scheduler"
)

// ServiceParameters es una estructura que encapsula los parametros
// de configuración de un nuevo servicio
type ServiceParameters struct {
	Id                     string
	Version                string
	ImageName              string
	ImageTag               string
	MinInstancesPerCluster map[string]int
}

// ServiceInstance contiene la información de una instancia de un servicio
type ServiceInstance struct {
	Id           string
	CreationDate time.Time
	Address      string
	Port         int
	Status       scheduler.ServiceInformationStatus // TODO No se debe depender de scheduler
	ClusterId    string
	ImageName    string
	ImageTag     string
}
