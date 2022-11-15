package types

import (
	"time"

	"github.com/galaxy-future/costpilot/internal/constants/cloud"
)

type (
	Granularity string

	MetricItem string
)

const (
	Monthly Granularity = "MONTHLY"
	Daily   Granularity = "DAILY"

	MetricItemCPUUtilization        MetricItem = "cpu.utilization"
	MetricItemMemoryUsedUtilization MetricItem = "memory.used.utilization"
)

type QueryAccountBillRequest struct {
	BillingCycle     string      `position:"Query" name:"BillingCycle"`
	BillingDate      string      `position:"Query" name:"BillingDate"`
	IsGroupByProduct bool        `position:"Query" name:"IsGroupByProduct"`
	Granularity      Granularity `position:"Query" name:"Granularity"`

	// ProductCode      string           `position:"Query" name:"ProductCode"`
	// PageNum          int `position:"Query" name:"PageNum"`
	// OwnerID          requests.Integer `position:"Query" name:"OwnerID"`
	// BillOwnerId      requests.Integer `position:"Query" name:"BillOwnerId"`
	// PageSize         int `position:"Query" name:"PageSize"`
}

type DataInQueryAccountBill struct {
	BillingCycle string                  `json:"BillingCycle" xml:"BillingCycle"`
	AccountID    string                  `json:"AccountID" xml:"AccountID"`
	TotalCount   int                     `json:"TotalCount" xml:"TotalCount"`
	AccountName  string                  `json:"AccountName" xml:"AccountName"`
	Items        ItemsInQueryAccountBill `json:"Items" xml:"Items"`
}
type ItemsInQueryAccountBill struct {
	Item []AccountBillItem `json:"Item" xml:"Item"`
}

type AccountBillItem struct {
	PipCode          PipCode                `json:"PipCode" xml:"PipCode"`
	ProductName      string                 `json:"ProductName" xml:"ProductName"`
	BillingDate      string                 `json:"BillingDate" xml:"BillingDate"`
	SubscriptionType cloud.SubscriptionType `json:"SubscriptionType" xml:"SubscriptionType"`
	Currency         string                 `json:"Currency" xml:"Currency"`
	PretaxAmount     float64                `json:"PretaxAmount" xml:"PretaxAmount"` // 应付金额
}

type ServiceType struct {
	ServiceTypeName string `json:"service_type_name"`
	ServiceTypeCode string `json:"service_type_code"`
	Abbreviation    string `json:"abbreviation"`
}

type DescribeMetricListRequest struct {
	MetricName         MetricItem
	Period             string
	StartTime, EndTime time.Time
}

type MetricSample struct {
	Timestamp         int64
	InstanceId        string
	Min, Max, Average float64
}

type DescribeMetricList struct {
	List []MetricSample
}

type (
	ResourceType   string
	RegionLanguage string
)

const (
	ResourceTypeInstance ResourceType = "instance"
	ResourceTypeDisk     ResourceType = "disk"

	RegionLanguageENUS RegionLanguage = "en-US"
	RegionLanguageZHCN RegionLanguage = "zh-CN"
)

type DescribeRegionsRequest struct {
	ResourceType ResourceType
	Language     RegionLanguage
}

type Region struct {
	RegionEndpoint string
	LocalName      string
	RegionId       string
}

type DescribeRegions struct {
	List []Region
}

type DescribeInstanceAttributeRequest struct {
	InstanceId string
}

type DescribeInstanceAttribute struct {
	InstanceId          string
	InstanceName        string
	RegionId            string
	HostName            string
	Status              string
	InstanceType        string
	InstanceNetworkType string
	SubscriptionType    cloud.SubscriptionType
	Memory              int32
	Cpu                 int32
	ImageId             string
	StoppedMode         string
	InternetChargeType  string
	PublicIpAddress     []string
	InnerIpAddress      []string
}

type DescribeInstanceBillRequest struct {
	// YYYY－MM
	BillingCycle string
	// ProductCode      string
	// ProductType      string
	// SubscriptionType string
	Granularity Granularity
	InstanceId  string
}

type ItemsInInstanceBill struct {
	BillingDate      string
	InstanceConfig   string
	InternetIP       string // 公网IP
	IntranetIP       string // 内网IP
	InstanceId       string
	Currency         string
	SubscriptionType cloud.SubscriptionType
	InstanceSpec     string
	Region           string
	ProductName      string
	ProductDetail    string
	ItemName         string // 项目名称
}

type DescribeInstanceBill struct {
	BillingCycle string                `json:"BillingCycle" xml:"BillingCycle"`
	AccountID    string                `json:"AccountID" xml:"AccountID"`
	TotalCount   int                   `json:"TotalCount" xml:"TotalCount"`
	AccountName  string                `json:"AccountName" xml:"AccountName"`
	Items        []ItemsInInstanceBill `json:"Items" xml:"Items"`
}

type QueryAvailableInstancesRequest struct {
	RegionId    string
	ProductCode string

	SubscriptionType cloud.SubscriptionType
	InstanceIdList   []string // max 100
}

type ItemAvailableInstance struct {
	InstanceId       string
	RegionId         string
	Status           string
	RenewStatus      string
	SubscriptionType cloud.SubscriptionType

	ProductCode string
}
type QueryAvailableInstances struct {
	TotalCount int
	List       []ItemAvailableInstance
}
