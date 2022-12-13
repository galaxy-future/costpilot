package huawei

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/galaxy-future/costpilot/tools/limiter"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	bss "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2"
	bssModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2/model"
	regionHuawei "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2/region"
	ces "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1"
	cesModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	cesRegion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/region"
	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	ecsModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	ecsRegion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/region"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	iamModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	iamRegion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
)

const (
	_bssRegion       = "cn-north-1"
	_chargingMode    = "charging_mode"
	_floating        = "floating"
	_instanceId      = "instance_id"
	_average         = "average"
	_namespaceSysECS = "SYS.ECS"
)

var huaweiMetric = map[types.MetricItem]string{
	types.MetricItemCPUUtilization:        "cpu_util",
	types.MetricItemMemoryUsedUtilization: "mem_usedPercent",
}

type HuaweiCloud struct {
	bssClientOpt *bss.BssClient
	iamClient    *iam.IamClient
	ecsClient    *ecs.EcsClient
	cesClient    *ces.CesClient
}

func New(AK, SK, region string) (*HuaweiCloud, error) {
	auth := global.NewCredentialsBuilder().
		WithAk(AK).
		WithSk(SK).
		Build()

	basicAuth := basic.NewCredentialsBuilder().
		WithAk(AK).
		WithSk(SK).
		Build()

	bssClientOpt := bss.NewBssClient(bss.BssClientBuilder().WithRegion(regionHuawei.ValueOf(_bssRegion)).WithCredential(auth).Build())

	iamClient := iam.NewIamClient(
		iam.IamClientBuilder().
			WithRegion(iamRegion.ValueOf(region)).
			WithCredential(auth).
			Build())

	ecsClient := ecs.NewEcsClient(
		ecs.EcsClientBuilder().
			WithRegion(ecsRegion.ValueOf(region)).
			WithCredential(basicAuth).
			Build())

	cesClient := ces.NewCesClient(
		ces.CesClientBuilder().
			WithRegion(cesRegion.ValueOf(region)).
			WithCredential(basicAuth).
			Build())

	return &HuaweiCloud{
		bssClientOpt: bssClientOpt,
		iamClient:    iamClient,
		ecsClient:    ecsClient,
		cesClient:    cesClient,
	}, nil
}

// ProviderType
func (*HuaweiCloud) ProviderType() cloud.Provider {
	return cloud.HuaweiCloud
}

func (p *HuaweiCloud) QueryAccountBill(ctx context.Context, param types.QueryAccountBillRequest) (result types.DataInQueryAccountBill, err error) {
	if param.Granularity == types.Daily {
		result, err = p.queryAccountBillByDate(ctx, param)
	}
	if param.Granularity == types.Monthly {
		result, err = p.queryAccountBillByMonth(ctx, param)
	}
	if err != nil {
		return types.DataInQueryAccountBill{}, err
	}

	return result, nil
}

func convQueryAccountBillByMonth(param types.QueryAccountBillRequest, response *bssModel.ShowCustomerMonthlySumResponse) []types.AccountBillItem {
	if response == nil || response.BillSums == nil {
		return []types.AccountBillItem{}
	}
	result := make([]types.AccountBillItem, 0, len(*response.BillSums))
	for _, v := range *response.BillSums {
		temp := types.AccountBillItem{
			PipCode:          convPipCode(tea.StringValue(v.ServiceTypeCode)),
			ProductName:      tea.StringValue(v.ServiceTypeName),
			SubscriptionType: convSubscriptionType(strconv.Itoa(int(tea.Int32Value(v.ChargingMode)))),
			Currency:         tea.StringValue(response.Currency),
			PretaxAmount:     tea.Float64Value(v.CashAmount),
		}
		result = append(result, temp)
	}
	return result
}

func convQueryAccountBill(response *bssModel.ListCustomerselfResourceRecordsResponse) []types.AccountBillItem {
	if response == nil {
		return []types.AccountBillItem{}
	}

	feeRecords := *response.FeeRecords
	result := make([]types.AccountBillItem, 0, len(feeRecords))
	for _, v := range feeRecords {
		standardPipCode := convPipCode(tea.StringValue(v.CloudServiceType))
		item := types.AccountBillItem{
			PipCode:          standardPipCode,
			ProductName:      tea.StringValue(v.CloudServiceTypeName),
			BillingDate:      tea.StringValue(v.BillDate),
			SubscriptionType: convSubscriptionType(*v.ChargeMode),
			Currency:         tea.StringValue(response.Currency),
			PretaxAmount:     tea.Float64Value(v.Amount),
		}
		result = append(result, item)
	}

	return result
}

func convSubscriptionType(chargeMode string) cloud.SubscriptionType {
	switch chargeMode {
	// 0：按需
	case "0":
		return cloud.PostPaid
	// 1:包年/包月
	case "1":
		return cloud.PrePaid
	}
	return "undefined"
}

func convPipCode(pipCode string) types.PipCode {
	switch pipCode {
	// 弹性云服务器
	case "hws.service.type.ec2":
		return types.ECS
	// 弹性公网IP
	case "hws.service.type.eip":
		return types.EIP
	// 对象存储服务
	case "hws.service.type.obs":
		return types.S3
		// NAT网关
	case "hws.service.type.natgateway":
		return types.NAT
		// 弹性文件服务
	case "hws.service.type.sfs":
		return types.NAS
		// 弹性负载均衡
	case "hws.service.type.elb":
		return types.SLB
		// 分布式缓存服务
	case "hws.service.type.dcs":
		return types.KVSTORE
		// 云桌面
	case "hws.service.type.vdi":
		return types.GWS
		// 企业网络部署规划设计服务
	case "hws.resource.type.pds.en":
		return types.CBN
	}
	return types.PipCode(pipCode)
}

func getIpInfoForECS(server ecsModel.ServerDetail) (fixedIps []string, floatingIps []string) {
	for _, addresses := range server.Addresses {
		for _, address := range addresses {
			if address.OSEXTIPStype != nil && address.OSEXTIPStype.Value() == _floating {
				floatingIps = append(floatingIps, address.Addr)
			} else {
				fixedIps = append(fixedIps, address.Addr)
			}
		}
	}
	return
}

func (p *HuaweiCloud) DescribeMetricList(ctx context.Context, param types.DescribeMetricListRequest) (types.DescribeMetricList, error) {
	request := &cesModel.BatchListMetricDataRequest{}
	dimensions := make([]cesModel.MetricsDimension, 0)
	if len(param.Filter.InstanceIds) == 1 {
		dimensions = append(dimensions, cesModel.MetricsDimension{
			Name:  _instanceId,
			Value: param.Filter.InstanceIds[0],
		})
	} else {
		return types.DescribeMetricList{}, fmt.Errorf("filter InstanceIds for metric incorrect")
	}
	metric := cesModel.MetricInfo{
		Namespace:  _namespaceSysECS,
		Dimensions: dimensions,
	}
	if hm, ok := huaweiMetric[param.MetricName]; ok {
		metric.MetricName = hm
	} else {
		return types.DescribeMetricList{}, fmt.Errorf("collect metric %s not supported for huawei", param.MetricName)
	}
	request.Body = &cesModel.BatchListMetricDataRequestBody{
		Metrics: []cesModel.MetricInfo{metric},
		Period:  param.Period,
		Filter:  _average,
		From:    param.StartTime.UnixMilli(),
		To:      param.EndTime.UnixMilli(),
	}
	response, err := p.cesClient.BatchListMetricData(request)
	if err != nil {
		return types.DescribeMetricList{}, err
	}
	ret := types.DescribeMetricList{}
	if response.HttpStatusCode != http.StatusOK {
		return ret, fmt.Errorf("httpcode %d", response.HttpStatusCode)
	}
	if response.Metrics == nil || len(*response.Metrics) == 0 {
		return ret, nil
	}
	for _, data := range *response.Metrics {
		var id string
		if len(*data.Dimensions) > 0 {
			id = (*data.Dimensions)[0].Value
		}
		for _, datapoint := range data.Datapoints {
			ret.List = append(ret.List, types.MetricSample{
				Timestamp:  datapoint.Timestamp,
				InstanceId: id,
				Average:    *datapoint.Average,
			})
		}
	}
	return ret, nil
}

func (p *HuaweiCloud) DescribeRegions(ctx context.Context, param types.DescribeRegionsRequest) (types.DescribeRegions, error) {
	request := &iamModel.KeystoneListRegionsRequest{}
	response, err := p.iamClient.KeystoneListRegions(request)
	if err != nil {
		return types.DescribeRegions{}, err
	}
	ret := types.DescribeRegions{}
	if response.HttpStatusCode != http.StatusOK {
		return ret, fmt.Errorf("httpcode %d", response.HttpStatusCode)
	}
	if response.Regions == nil || len(*response.Regions) == 0 {
		return ret, nil
	}
	for _, r := range *response.Regions {
		localName := r.Locales.ZhCn
		if param.Language == types.RegionLanguageENUS {
			localName = r.Locales.EnUs
		}
		ret.List = append(ret.List, types.ItemRegion{
			LocalName: localName,
			RegionId:  r.Id,
		})
	}
	return ret, nil
}

func (p *HuaweiCloud) DescribeInstances(ctx context.Context, param types.DescribeInstancesRequest) (types.DescribeInstances, error) {
	request := &ecsModel.ListServersDetailsRequest{}
	response, err := p.ecsClient.ListServersDetails(request)
	if err != nil {
		return types.DescribeInstances{}, err
	}
	ret := types.DescribeInstances{}
	if response.HttpStatusCode != http.StatusOK {
		return ret, nil
	}
	if response.Servers == nil || len(*response.Servers) == 0 {
		return ret, nil
	}
	filterIdMap := make(map[string]bool)
	for _, id := range param.InstanceIds {
		filterIdMap[id] = true
	}
	for _, server := range *response.Servers {
		if !filterIdMap[server.Id] && len(filterIdMap) > 0 {
			continue
		}
		fixedIps, floatingIps := getIpInfoForECS(server)
		chargeMode, _ := server.Metadata[_chargingMode]
		ret.List = append(ret.List, types.ItemDescribeInstance{
			InstanceId:       server.Id,
			InstanceName:     server.Name,
			SubscriptionType: convSubscriptionType(chargeMode),
			InnerIpAddress:   fixedIps,
			PublicIpAddress:  floatingIps,
		})
	}
	ret.TotalCount = len(ret.List)
	return ret, nil
}

func (p *HuaweiCloud) DescribeInstanceBill(ctx context.Context, param types.DescribeInstanceBillRequest, isAll bool) (types.DescribeInstanceBill, error) {
	return types.DescribeInstanceBill{}, nil
}

func (p *HuaweiCloud) QueryAvailableInstances(ctx context.Context, param types.QueryAvailableInstancesRequest) (types.QueryAvailableInstances, error) {
	return types.QueryAvailableInstances{}, nil
}

func (p *HuaweiCloud) queryAccountBillByMonth(ctx context.Context, param types.QueryAccountBillRequest) (result types.DataInQueryAccountBill, err error) {

	billItems := make([]types.AccountBillItem, 0)
	request := &bssModel.ShowCustomerMonthlySumRequest{}
	request.BillCycle = param.BillingCycle
	pageNum := int32(0)
	request.Offset = tea.Int32(0)
	request.Limit = tea.Int32(10)
	response := new(bssModel.ShowCustomerMonthlySumResponse)
	if param.IsGroupByProduct {
		for {
			limiter := limiter.Limiters.GetLimiter(p.ProviderType().String()+"-"+"ShowCustomerMonthlySum", 9)
			limiter.Take()
			response, err = p.bssClientOpt.ShowCustomerMonthlySum(request)

			if err != nil {
				return types.DataInQueryAccountBill{}, err
			}
			totalCount := response.TotalCount
			if len(billItems) == 0 {
				billItems = make([]types.AccountBillItem, 0, *totalCount)
			}
			billItems = append(billItems, convQueryAccountBillByMonth(param, response)...)
			if len(billItems) >= int(*totalCount) {
				break
			}
			pageNum += *request.Limit
			request.Offset = &pageNum
		}
		result = types.DataInQueryAccountBill{
			BillingCycle: param.BillingCycle,
			TotalCount:   int(tea.Int32Value(response.TotalCount)),
			Items:        types.ItemsInQueryAccountBill{Item: billItems},
		}
	} else {
		limiter := limiter.Limiters.GetLimiter(p.ProviderType().String()+"-"+"ShowCustomerMonthlySum", 9)
		limiter.Take()
		response, err = p.bssClientOpt.ShowCustomerMonthlySum(request)
		if err != nil {
			return types.DataInQueryAccountBill{}, err
		}
		result = types.DataInQueryAccountBill{
			BillingCycle: param.BillingCycle,
			TotalCount:   int(tea.Int32Value(response.TotalCount)),
			Items: types.ItemsInQueryAccountBill{
				Item: []types.AccountBillItem{
					{
						Currency:     tea.StringValue(response.Currency),
						PretaxAmount: tea.Float64Value(response.CashAmount),
					},
				},
			},
		}
	}

	return result, nil
}
func (p *HuaweiCloud) queryAccountBillByDate(ctx context.Context, param types.QueryAccountBillRequest) (result types.DataInQueryAccountBill, err error) {
	billItems := make([]types.AccountBillItem, 0)
	request := &bssModel.ListCustomerselfResourceRecordsRequest{}
	request.Cycle = param.BillingCycle
	request.BillDateBegin = tea.String(param.BillingDate)
	request.BillDateEnd = tea.String(param.BillingDate)
	includeZeroRecordRequest := false
	request.IncludeZeroRecord = &includeZeroRecordRequest
	request.Offset = tea.Int32(0)
	request.Limit = tea.Int32(10)
	pageNum := int32(0)
	response := new(bssModel.ListCustomerselfResourceRecordsResponse)
	for {
		limiter := limiter.Limiters.GetLimiter(p.ProviderType().String()+"-"+"ListCustomerselfResourceRecords", 9)
		limiter.Take()
		response, err = p.bssClientOpt.ListCustomerselfResourceRecords(request)

		if err != nil {
			return types.DataInQueryAccountBill{}, err
		}
		totalCount := response.TotalCount
		if len(billItems) == 0 {
			billItems = make([]types.AccountBillItem, 0, *totalCount)
		}
		billItems = append(billItems, convQueryAccountBill(response)...)
		if len(billItems) >= int(*totalCount) {
			break
		}
		pageNum += tea.Int32Value(request.Limit)
		request.Offset = &pageNum
	}
	tempMap := make(map[string]*types.AccountBillItem, len(billItems))
	if param.IsGroupByProduct {
		for _, v := range billItems {
			if val, ok := tempMap[v.ProductName+v.SubscriptionType.String()]; ok {
				val.PretaxAmount += v.PretaxAmount
			} else {
				tempMap[v.ProductName+v.SubscriptionType.String()] = &types.AccountBillItem{
					PipCode:          v.PipCode,
					ProductName:      v.ProductName,
					BillingDate:      v.BillingDate,
					SubscriptionType: v.SubscriptionType,
					Currency:         v.Currency,
					PretaxAmount:     v.PretaxAmount,
				}
			}
		}
	} else {
		var totalCost float64
		var currency string
		for _, v := range billItems {
			totalCost += v.PretaxAmount
			if currency == "" {
				currency = v.Currency
			}
		}
		return types.DataInQueryAccountBill{
			BillingCycle: param.BillingCycle,
			AccountID:    "",
			TotalCount:   1,
			AccountName:  "",
			Items: types.ItemsInQueryAccountBill{
				Item: []types.AccountBillItem{
					{
						BillingDate:  param.BillingDate,
						Currency:     currency,
						PretaxAmount: totalCost,
					},
				},
			},
		}, nil
	}
	resultItems := make([]types.AccountBillItem, 0, len(tempMap))
	for _, v := range tempMap {
		resultItems = append(resultItems, *v)
	}
	result = types.DataInQueryAccountBill{
		BillingCycle: param.BillingCycle,
		AccountID:    "",
		TotalCount:   len(resultItems),
		AccountName:  "",
		Items: types.ItemsInQueryAccountBill{
			Item: resultItems,
		},
	}

	return result, nil
}
