package types

import (
	"time"
)

type Instance struct {
	Id      string `json:"id"`
	Status  string `json:"status,omitempty"`
	Cluster string `json:"cluster,omitempty"`
	Address string `json:"address"`
}

type ServiceVersion struct {
	Version      string     `json:"version,omitempty"`
	CreationDate *time.Time `json:"creation_time,omitempty"`
	ImageName    string     `json:"image_name,omitempty"`
	ImageTag     string     `json:"image_tag,omitempty"`
	Instances    []Instance `json:"instances,omitempty"`
}

type Service struct {
	Id           string           `json:"id"`
	CreationDate *time.Time       `json:"creation_time,omitempty"`
	Versions     []ServiceVersion `json:"versions,omitempty"`
}
