package types

import "github.com/galaxy-future/costpilot/internal/constants/cloud"

type CloudAccount struct {
	Provider cloud.Provider `json:"provider" yaml:"provider"`
	AK       string         `json:"ak" yaml:"ak"`
	SK       string         `json:"sk" yaml:"sk"`
	RegionID string         `json:"region_id" yaml:"region_id"`
	Name     string         `json:"name" yaml:"name"`
}
