package alibaba

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	bssopenapiV3 "github.com/alibabacloud-go/bssopenapi-20171214/v3/client"
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
	bssClientNew *bssopenapiV3.Client
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
	bssClientNew, err := bssopenapiV3.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(AK),
		AccessKeySecret: tea.String(SK),
		Endpoint:        tea.String("business.aliyuncs.com"),
	})
	if err != nil {
		return nil, err
	}

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
		bssClientNew: bssClientNew,
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
			SubscriptionType: convSubscriptionTypeAliyunToCloud(v.SubscriptionType),
			Currency:         v.Currency,
			PretaxAmount:     v.PretaxAmount,
		}
		result = append(result, item)
	}

	return result
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
		Length:     tea.String("100"),
	}
	var allDataList []*Datapoint
	page := 0
	for {
		log.Printf("I! Fecth metrics for page[%d]\n", page)
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
		if len(dataList) > 0 {
			allDataList = append(allDataList, dataList...)
		}
		if response.Body.NextToken == nil {
			break
		} else {
			request.NextToken = tea.String(*response.Body.NextToken)
			page++
		}
	}

	ret := types.DescribeMetricList{List: make([]types.MetricSample, 0, len(allDataList))}
	for _, datapoint := range allDataList {
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
		SubscriptionType:    convSubscriptionTypeAliyunToCloud(*responseBody.InstanceChargeType),
		Memory:              *responseBody.Memory,
		Cpu:                 *responseBody.Cpu,
		ImageId:             *responseBody.ImageId,
		StoppedMode:         *responseBody.StoppedMode,
		InternetChargeType:  *responseBody.InternetChargeType,
		RegionId:            *responseBody.RegionId,
	}, nil
}

// DescribeInstanceBill 实例账单是根据账单数据拆分生成，一般会有一天延迟。
func (p *AlibabaCloud) DescribeInstanceBill(_ context.Context, param types.DescribeInstanceBillRequest, isAll bool) (types.DescribeInstanceBill, error) {
	if param.BillingCycle == "" {
		return types.DescribeInstanceBill{}, errors.New("BillingCycle empty")
	}
	request := &bssopenapiV3.DescribeInstanceBillRequest{
		BillingCycle: tea.String(param.BillingCycle),
		MaxResults:   tea.Int32(_maxLimit), // alibaba cloud max limit
	}
	if param.InstanceId != "" {
		request.InstanceID = tea.String(param.InstanceId)
	}
	if param.Granularity != "" {
		granularity := tea.ToString(param.Granularity)
		request.Granularity = &granularity
	}
	var (
		billItems []types.ItemsInInstanceBill
		respData  *bssopenapiV3.DescribeInstanceBillResponseBodyData
	)
	for {
		response, err := p.bssClientNew.DescribeInstanceBill(request)
		if err != nil {
			return types.DescribeInstanceBill{}, err
		}
		if *response.StatusCode != http.StatusOK {
			return types.DescribeInstanceBill{}, fmt.Errorf("httpcode %d", *response.StatusCode)
		}
		respData = response.Body.Data
		totalCount := *respData.TotalCount
		if len(billItems) == 0 && isAll {
			billItems = make([]types.ItemsInInstanceBill, 0, totalCount)
		}
		billItems = append(billItems, convInstanceBill(respData)...)
		if !isAll {
			break
		}
		if len(billItems) >= int(totalCount) {
			break
		}
		request.NextToken = respData.NextToken
	}

	result := types.DescribeInstanceBill{
		BillingCycle: *respData.BillingCycle,
		AccountID:    *respData.AccountID,
		TotalCount:   len(billItems),
		AccountName:  *respData.AccountName,
		Items:        billItems,
	}
	return result, nil
}

func convInstanceBill(respData *bssopenapiV3.DescribeInstanceBillResponseBodyData) []types.ItemsInInstanceBill {
	if respData == nil || len(respData.Items) == 0 {
		return []types.ItemsInInstanceBill{}
	}
	result := make([]types.ItemsInInstanceBill, 0, len(respData.Items))

	for _, item := range respData.Items {
		result = append(result, types.ItemsInInstanceBill{
			BillingDate:      *item.BillingDate,
			InstanceConfig:   *item.InstanceConfig,
			InternetIP:       *item.InternetIP,
			IntranetIP:       *item.IntranetIP,
			InstanceId:       *item.InstanceID,
			Currency:         *item.Currency,
			SubscriptionType: convSubscriptionTypeAliyunToCloud(*item.SubscriptionType),
			InstanceSpec:     *item.InstanceSpec,
			Region:           *item.Region,
			ProductName:      *item.ProductName,
			ProductDetail:    *item.ProductDetail,
			ItemName:         *item.ItemName,
		})
	}
	return result
}
func convSubscriptionTypeAliyunToCloud(subscriptionType string) cloud.SubscriptionType {
	switch subscriptionType {
	case "Subscription":
		return cloud.PrePaid
	case "PayAsYouGo":
		return cloud.PostPaid
	default:
		return cloud.Undefined
	}
}

func convSubscriptionTypeCloudToAliyun(st cloud.SubscriptionType) string {
	switch st {
	case cloud.PrePaid:
		return "Subscription"
	case cloud.PostPaid:
		return "PayAsYouGo"
	default:
		return ""
	}
}

// alicloud: InstanceIDs in QueryAvailableInstancesRequest are max to 100
func (p *AlibabaCloud) QueryAvailableInstances(ctx context.Context, param types.QueryAvailableInstancesRequest) (types.QueryAvailableInstances, error) {
	if len(param.InstanceIdList) <= 100 {
		return p.queryAvailableInstancesByPage(ctx, param)
	} else {
		total := len(param.InstanceIdList)
		var instanceList []types.ItemAvailableInstance
		for i := 0; i < total; i += 100 {
			endIdx := i + 100
			if endIdx > total {
				endIdx = total
			}
			log.Printf("!I GetInstanceList page[%d], for total[%d]", i/100, total)
			pageResult, err := p.queryAvailableInstancesByPage(ctx, types.QueryAvailableInstancesRequest{
				RegionId:         param.RegionId,
				ProductCode:      param.ProductCode,
				SubscriptionType: param.SubscriptionType,
				InstanceIdList:   param.InstanceIdList[i:endIdx],
			})
			if err != nil {
				return types.QueryAvailableInstances{}, err
			}
			instanceList = append(instanceList, pageResult.List...)
		}
		result := types.QueryAvailableInstances{TotalCount: len(instanceList), List: instanceList}
		return result, nil
	}
}

func (p *AlibabaCloud) queryAvailableInstancesByPage(_ context.Context, param types.QueryAvailableInstancesRequest) (types.QueryAvailableInstances, error) {
	request := &bssopenapiV3.QueryAvailableInstancesRequest{}
	if len(param.RegionId) > 0 {
		if len(param.ProductCode) == 0 {
			return types.QueryAvailableInstances{}, errors.New("The parameter productCode cannot be blank when the region is not blank")
		}
		request.Region = tea.String(param.RegionId)
	}

	if len(param.InstanceIdList) > 100 {
		return types.QueryAvailableInstances{}, errors.Errorf("InstanceIDs in QueryAvailableInstancesRequest are max to 100, current: %d", len(param.InstanceIdList))
	}

	if len(param.InstanceIdList) > 0 {
		request.InstanceIDs = tea.String(strings.Join(param.InstanceIdList, ","))
	}

	if param.SubscriptionType != "" {
		st := convSubscriptionTypeCloudToAliyun(param.SubscriptionType)
		request.SubscriptionType = &st
	}
	var pageNum int32 = 1
	request.PageNum = &pageNum
	request.PageSize = &_maxLimit // alibaba cloud max limit

	var (
		instanceList []types.ItemAvailableInstance
		respData     *bssopenapiV3.QueryAvailableInstancesResponseBodyData
	)
	for {
		response, err := p.bssClientNew.QueryAvailableInstances(request)
		if err != nil {
			return types.QueryAvailableInstances{}, err
		}
		if *response.StatusCode != http.StatusOK {
			return types.QueryAvailableInstances{}, fmt.Errorf("httpcode %d", *response.StatusCode)
		}

		if response.Body.Success == nil || (!*response.Body.Success) {
			fmt.Printf("QueryAvailableInstances err: %v\n", *response.Body.Message)
			return types.QueryAvailableInstances{}, fmt.Errorf("QueryAvailableInstances err: %v", *response.Body.Message)
		}

		respData = response.Body.Data
		total := respData.TotalCount
		if len(instanceList) == 0 {
			instanceList = make([]types.ItemAvailableInstance, 0, *total)
		}
		instanceList = append(instanceList, convAvailableInstances(respData)...)
		if int32(len(instanceList)) >= *total {
			break
		}
		pageNum++
		request.PageNum = &pageNum
	}

	result := types.QueryAvailableInstances{TotalCount: len(instanceList), List: instanceList}

	return result, nil
}

func convAvailableInstances(respData *bssopenapiV3.QueryAvailableInstancesResponseBodyData) []types.ItemAvailableInstance {
	if respData == nil || len(respData.InstanceList) == 0 {
		return []types.ItemAvailableInstance{}
	}
	result := make([]types.ItemAvailableInstance, 0, len(respData.InstanceList))
	for _, item := range respData.InstanceList {
		i := types.ItemAvailableInstance{
			InstanceId:       *item.InstanceID,
			RegionId:         *item.Region,
			Status:           *item.Status,
			RenewStatus:      *item.RenewStatus,
			SubscriptionType: convSubscriptionTypeAliyunToCloud(*item.SubscriptionType),
			ProductCode:      *item.ProductCode,
		}
		result = append(result, i)
	}

	return result
}
