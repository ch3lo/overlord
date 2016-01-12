package types

import "time"

type Instance struct {
	Id           string     `json:"id"`
	CreationDate *time.Time `json:"creation_time,omitempty"`
	Status       string     `json:"status,omitempty"`
	Cluster      string     `json:"cluster,omitempty"`
	Address      string     `json:"address"`
}

type AppMajorVersion struct {
	Version      string                  `json:"version,omitempty"`
	CreationDate *time.Time              `json:"creation_time,omitempty"`
	ImageName    string                  `json:"image_name,omitempty"`
	ImageTag     string                  `json:"image_tag,omitempty"`
	Instances    []Instance              `json:"instances,omitempty"`
	ClusterCheck map[string]ClusterCheck `json:"cluster_check"`
}

type Application struct {
	Id           string            `json:"id"`
	CreationDate *time.Time        `json:"creation_time,omitempty"`
	Versions     []AppMajorVersion `json:"versions,omitempty"`
}

type ClusterCheck struct {
	Instances int `json:"instances,omitempty"`
}

type ConstraintMapper struct {
	ImageName    string                  `json:"image_name,omitempty", valid:"alphanum"`
	ClusterCheck map[string]ClusterCheck `json:"cluster_check"`
}

type AppRequest struct {
	AppID        string           `json:"app_id", valid:"alphanum,required"`
	MajorVersion string           `json:"app_major_version", valid:"alphanum,required"`
	Constraints  ConstraintMapper `json:"constraints"`
}
