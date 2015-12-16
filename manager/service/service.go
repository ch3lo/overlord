package service

import "time"

// Parameters es una estructura que encapsula los parametros
// de configuración de un nuevo servicio
type Parameters struct {
	ID                     string
	Version                string
	ImageName              string
	ImageTag               string
	MinInstancesPerCluster map[string]int
}

// Instance contiene la información de una instancia de un servicio
type Instance struct {
	ID           string
	CreationDate time.Time
	Address      string
	Port         int
	Healthy      bool
	ClusterID    string
	ImageName    string
	ImageTag     string
}
