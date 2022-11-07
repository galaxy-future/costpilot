package alibaba

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cms "github.com/alibabacloud-go/cms-20190101/v8/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs "github.com/alibabacloud-go/ecs-20140526/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/pkg/errors"
)

type AlibabaCloud struct {
	bssClientOpt *bssopenapi.Client
	cmsClient    *cms.Client
	ecsClient    *ecs.Client
}

var (
	metricNameMap = map[types.MetricItem]string{
		types.MetricItemCPUUtilization:        "CPUUtilization",
		types.MetricItemMemoryUsedUtilization: "memory_usedutilization",
	}

	resourceTypeMap = map[types.ResourceType]string{
		types.ResourceTypeInstance: "Instance",
		types.ResourceTypeDisk:     "disk",
	}

	regionLanguageMap = map[types.RegionLanguage]string{
		types.RegionLanguageZHCN: "zh-CN",
		types.RegionLanguageENUS: "en-US",
	}
)

func New(AK, SK, region string) (*AlibabaCloud, error) {
	bssClientOpt, err := bssopenapi.NewClientWithOptions(region, sdk.NewConfig().WithTimeout(10*time.Second), credentials.NewAccessKeyCredential(AK, SK))
	if err != nil {
		return nil, err
	}

	config := &openapi.Config{
		AccessKeyId:     tea.String(AK),
		AccessKeySecret: tea.String(SK),
	}
	config.Endpoint = tea.String(_cmsEndPoint)
	cmsClient, err := cms.NewClient(config)
	if err != nil {
		return nil, err
	}

	ecsClient, err := ecs.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(AK),
		AccessKeySecret: tea.String(SK),
		Endpoint:        tea.String(_ecsEndPoint),
	})
	if err != nil {
		return nil, err
	}

	return &AlibabaCloud{
		bssClientOpt: bssClientOpt,
		cmsClient:    cmsClient,
		ecsClient:    ecsClient,
	}, nil
}

// ProviderType
func (*AlibabaCloud) ProviderType() cloud.Provider {
	return cloud.AlibabaCloud
}

// QueryAccountBill
func (p *AlibabaCloud) QueryAccountBill(ctx context.Context, param types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error) {
	var err error
	billItems := make([]types.AccountBillItem, 0)
	request := bssopenapi.CreateQueryAccountBillRequest()
	request.Scheme = "https"
	request.BillingDate = param.BillingDate
	request.BillingCycle = param.BillingCycle
	request.IsGroupByProduct = requests.NewBoolean(param.IsGroupByProduct)
	request.Granularity = tea.ToString(param.Granularity)
	pageNum := 1
	request.PageNum = requests.NewInteger(pageNum)
	request.PageSize = requests.NewInteger(300) // alibaba cloud max limit
	response := bssopenapi.CreateQueryAccountBillResponse()
	for {
		response, err = p.bssClientOpt.QueryAccountBill(request)
		if err != nil {
			return types.DataInQueryAccountBill{}, err
		}
		totalCount := response.Data.TotalCount
		if len(billItems) == 0 {
			billItems = make([]types.AccountBillItem, 0, totalCount)
		}
		billItems = append(billItems, convQueryAccountBill(response)...)
		if len(billItems) >= totalCount {
			break
		}
		pageNum++
		request.PageNum = requests.NewInteger(pageNum)
	}
	result := types.DataInQueryAccountBill{
		BillingCycle: response.Data.BillingCycle,
		AccountID:    response.Data.AccountID,
		TotalCount:   len(billItems),
		AccountName:  response.Data.AccountName,
		Items: types.ItemsInQueryAccountBill{
			Item: billItems,
		},
	}

	return result, nil
}

// convQueryAccountBill
func convQueryAccountBill(response *bssopenapi.QueryAccountBillResponse) []types.AccountBillItem {
	if response == nil {
		return nil
	}
	result := make([]types.AccountBillItem, 0, len(response.Data.Items.Item))
	for _, v := range response.Data.Items.Item {
		standardPipCode := convPipCode(v.PipCode)
		item := types.AccountBillItem{
			PipCode:          standardPipCode,
			ProductName:      convProductName(standardPipCode, v.ProductName),
			BillingDate:      v.BillingDate, // has date when Granularity=DAILY
			SubscriptionType: convSubscriptionType(v.SubscriptionType),
			Currency:         v.Currency,
			PretaxAmount:     v.PretaxAmount,
		}
		result = append(result, item)
	}

	return result
}

func convSubscriptionType(subscriptionType string) cloud.SubscriptionType {
	switch subscriptionType {
	case "PrePaid":
		return cloud.PrePaid
	case "PostPaid":
		return cloud.PostPaid
	}
	return "undefined"
}

func convPipCode(pipCode string) types.PipCode {
	// switch pipCode {
	// case "oss":
	//	return types.S3
	// }
	// 暂不启用转换，直接返回
	return types.PipCode(pipCode)
}

func convProductName(pipCode types.PipCode, defaultName ...string) string {
	// name := types.PidCode2Name(pipCode)
	// if name == types.Undefined && len(defaultName) != 0 {
	//	return defaultName[0]
	// }
	// 暂不启用转换，直接返回
	return defaultName[0]
}

func (p *AlibabaCloud) DescribeMetricList(_ context.Context, param types.DescribeMetricListRequest) (types.DescribeMetricList, error) {
	type Datapoint struct {
		InstanceId string  `json:"instanceId"`
		Timestamp  int64   `json:"timestamp"`
		UserId     string  `json:"userId"`
		Minimum    float64 `json:"Minimum"`
		Maximum    float64 `json:"Maximum"`
		Average    float64 `json:"Average"`
	}
	metricName, ok := metricNameMap[param.MetricName]
	if !ok {
		return types.DescribeMetricList{}, errors.New("unknown metric name")
	}
	request := &cms.DescribeMetricListRequest{
		Namespace:  tea.String("acs_ecs_dashboard"),
		MetricName: tea.String(metricName),
		Period:     tea.String(param.Period),
		StartTime:  tea.String(param.StartTime.Format("2006-01-02T15:04:05Z")),
		EndTime:    tea.String(param.EndTime.Format("2006-01-02T15:04:05Z")),
	}
	response, err := p.cmsClient.DescribeMetricList(request)
	if err != nil {
		return types.DescribeMetricList{}, err
	}
	if *response.StatusCode != http.StatusOK {
		return types.DescribeMetricList{}, fmt.Errorf("httpcode %d", *response.StatusCode)
	}
	if response.Body == nil || response.Body.Datapoints == nil {
		return types.DescribeMetricList{}, nil
	}
	dataStr := *response.Body.Datapoints
	var dataList []*Datapoint
	if err := json.Unmarshal([]byte(dataStr), &dataList); err != nil {
		return types.DescribeMetricList{}, nil
	}

	ret := types.DescribeMetricList{List: make([]types.MetricSample, 0, len(dataList))}

	for _, datapoint := range dataList {
		d := types.MetricSample{
			InstanceId: datapoint.InstanceId,
			Min:        datapoint.Minimum,
			Max:        datapoint.Maximum,
			Average:    datapoint.Average,
			Timestamp:  datapoint.Timestamp,
		}
		ret.List = append(ret.List, d)
	}

	return ret, nil
}

func (p *AlibabaCloud) DescribeRegions(_ context.Context, param types.DescribeRegionsRequest) (types.DescribeRegions, error) {
	resourceType, ok := resourceTypeMap[param.ResourceType]
	if !ok {
		return types.DescribeRegions{}, errors.New("unknown resource type")
	}

	lang, ok := regionLanguageMap[param.Language]
	if !ok {
		return types.DescribeRegions{}, errors.New("unknown region language")
	}

	request := &ecs.DescribeRegionsRequest{
		ResourceType:   tea.String(resourceType),
		AcceptLanguage: tea.String(lang),
	}

	response, err := p.ecsClient.DescribeRegionsWithOptions(request, &util.RuntimeOptions{})
	if err != nil {
		return types.DescribeRegions{}, err
	}
	ret := types.DescribeRegions{}
	if *response.StatusCode != http.StatusOK {
		return ret, fmt.Errorf("httpcode %d", *response.StatusCode)
	}
	if response.Body.Regions == nil || len(response.Body.Regions.Region) == 0 {
		return ret, nil
	}
	for _, r := range response.Body.Regions.Region {
		ret.List = append(ret.List, types.Region{
			RegionEndpoint: *r.RegionEndpoint,
			LocalName:      *r.LocalName,
			RegionId:       *r.RegionId,
		})
	}
	return ret, nil
}

func (p *AlibabaCloud) DescribeInstanceAttribute(_ context.Context, param types.DescribeInstanceAttributeRequest) (types.DescribeInstanceAttribute, error) {
	response, err := p.ecsClient.DescribeInstanceAttribute(&ecs.DescribeInstanceAttributeRequest{
		InstanceId: &param.InstanceId,
	})
	if err != nil {
		return types.DescribeInstanceAttribute{}, err
	}
	if *response.StatusCode != http.StatusOK {
		return types.DescribeInstanceAttribute{}, fmt.Errorf("httpcode %d", *response.StatusCode)
	}
	responseBody := response.Body

	return types.DescribeInstanceAttribute{
		InstanceId:          *responseBody.InstanceId,
		InstanceName:        *responseBody.InstanceName,
		HostName:            *responseBody.HostName,
		Status:              *responseBody.Status,
		InstanceType:        *responseBody.InstanceType,
		InstanceNetworkType: *responseBody.InstanceNetworkType,
		SubscriptionType:    convSubscriptionType(*responseBody.InstanceChargeType),
		Memory:              *responseBody.Memory,
		Cpu:                 *responseBody.Cpu,
		ImageId:             *responseBody.ImageId,
		StoppedMode:         *responseBody.StoppedMode,
		InternetChargeType:  *responseBody.InternetChargeType,
		RegionId:            *responseBody.RegionId,
	}, nil
}
