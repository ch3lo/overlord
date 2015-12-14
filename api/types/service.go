package types

import (
	"time"
)

type Instance struct {
	Id           string     `json:"id"`
	CreationDate *time.Time `json:"creation_time,omitempty"`
	Status       string     `json:"status,omitempty"`
	Cluster      string     `json:"cluster,omitempty"`
	Address      string     `json:"address"`
}

type ClusterCheck struct {
	Instaces int `json:"instances,omitempty"`
}

type ServiceManager struct {
	Version      string                  `json:"version,omitempty"`
	CreationDate *time.Time              `json:"creation_time,omitempty"`
	ImageName    string                  `json:"image_name,omitempty"`
	ImageTag     string                  `json:"image_tag,omitempty"`
	Instances    []Instance              `json:"instances,omitempty"`
	ClusterCheck map[string]ClusterCheck `json:"cluster_check,omitempty"`
}

type ServiceGroup struct {
	Id           string           `json:"id"`
	CreationDate *time.Time       `json:"creation_time,omitempty"`
	Managers     []ServiceManager `json:"managers,omitempty"`
}
