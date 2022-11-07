package data

import (
	"github.com/galaxy-future/costpilot/internal/constants/cloud"
)

type InstanceCpuUtilization struct {
	InstanceId      string
	UsedUtilization float64 // CPU使用率
}

type InstanceMemoryUtilization struct {
	InstanceId      string
	UsedUtilization float64 // 内存使用率
}

type DailyCpuUtilization struct {
	Provider    cloud.Provider
	Day         string                   `json:"day"` // 20220101
	Utilization []InstanceCpuUtilization `json:"utilization"`
}

type DailyMemoryUtilization struct {
	Provider    cloud.Provider
	Day         string                      `json:"day"` // 20220101
	Utilization []InstanceMemoryUtilization `json:"utilization"`
}

type InstanceDetail struct {
	Provider         cloud.Provider
	InstanceId       string
	RegionId         string
	RegionName       string
	SubscriptionType cloud.SubscriptionType
}
