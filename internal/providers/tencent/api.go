package tencent

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/galaxy-future/costpilot/tools/limiter"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/pkg/errors"
	billing "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/billing/v20180709"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
)

type TencentCloud struct {
	billingClient *billing.Client
	cvmClient     *cvm.Client
	monitorClient *monitor.Client
}

func New(ak, sk, regionId string) (*TencentCloud, error) {
	credential := common.NewCredential(ak, sk)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = _billingEndpoint
	billingClient, err := billing.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	cvmCP := profile.NewClientProfile()
	cvmCP.HttpProfile.Endpoint = _cvmEndPoint
	cvmClient, err := cvm.NewClient(credential, regionId, cvmCP)
	if err != nil {
		return nil, err
	}

	monitorCP := profile.NewClientProfile()
	monitorCP.HttpProfile.Endpoint = _monitorEndPoint
	monitorClient, err := monitor.NewClient(credential, regionId, monitorCP)
	if err != nil {
		return nil, err
	}

	return &TencentCloud{
		billingClient: billingClient,
		cvmClient:     cvmClient,
		monitorClient: monitorClient,
	}, nil
}

// ProviderType
func (*TencentCloud) ProviderType() cloud.Provider {
	return cloud.TencentCloud
}

// QueryAccountBill
func (p *TencentCloud) QueryAccountBill(_ context.Context, param types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error) {
	var (
		needNum int64  = 0
		offset  uint64 = 0
	)

	request := billing.NewDescribeBillDetailRequest()
	request.NeedRecordNum = &needNum
	request.Limit = &_maxPageSize
	request.Offset = &offset
	request.NeedRecordNum = common.Int64Ptr(1) // 1 for needing total ,0 for not

	// 根据不同的账单周期粒度组装数据
	switch param.Granularity {
	case types.Monthly:
		request.Month = &param.BillingCycle

	case types.Daily:
		beginTime, endTime, err := parseDateStartEndTime(param.BillingDate)
		if err != nil {
			return types.DataInQueryAccountBill{}, err
		}
		request.BeginTime = &beginTime
		request.EndTime = &endTime

	default:
		return types.DataInQueryAccountBill{}, errors.New("Unknown Granularity")
	}

	// 分页直到获取全部
	var allBillList []*billing.BillDetail
	for {
		limiter := limiter.Limiters.GetLimiter(p.ProviderType().String()+"-"+"DescribeBillDetail", 3)
		limiter.Take()
		response, err := p.billingClient.DescribeBillDetail(request)
		if err != nil {
			return types.DataInQueryAccountBill{}, err
		}
		if response == nil || response.Response == nil || response.Response.Total == nil {
			break
		}
		totalCount := response.Response.Total
		if len(allBillList) == 0 {
			allBillList = make([]*billing.BillDetail, 0, *totalCount)
		}
		allBillList = append(allBillList, response.Response.DetailSet...)
		if uint64(len(allBillList)) >= *totalCount {
			break
		}
		offset += _maxPageSize
		request.Offset = &offset
	}

	itemList, err := convQueryAccountBill(param, allBillList)
	if err != nil {
		return types.DataInQueryAccountBill{}, err
	}
	return types.DataInQueryAccountBill{
		BillingCycle: param.BillingCycle,
		TotalCount:   len(itemList),
		Items:        types.ItemsInQueryAccountBill{Item: itemList},
	}, nil
}

func convQueryAccountBill(param types.QueryAccountBillRequest, billList []*billing.BillDetail) ([]types.AccountBillItem, error) {
	if billList == nil {
		return nil, errors.New("invalid bill list")
	}
	if len(billList) == 0 {
		return []types.AccountBillItem{}, nil
	}
	billingItem := make([]types.AccountBillItem, 0)
	costMap := make(map[string]float64)
	checkMap := make(map[string]bool)
	var totalCost float64
	currency := convCurrency(tea.StringValue(billList[0].ComponentSet[0].PriceUnit))
	for _, item := range billList {
		costMap[tea.StringValue(item.BusinessCodeName)+tea.StringValue(item.PayModeName)] += sumComponentSet(item.ComponentSet)
	}
	if param.IsGroupByProduct {
		for _, item := range billList {
			if checkMap[tea.StringValue(item.BusinessCodeName)+tea.StringValue(item.PayModeName)] {
				continue
			}
			temp := types.AccountBillItem{
				PipCode:          types.PipCode(tea.StringValue(item.BusinessCode)),
				ProductName:      tea.StringValue(item.BusinessCodeName),
				BillingDate:      "",
				SubscriptionType: convSubscriptionType(tea.StringValue(item.PayModeName)),
				Currency:         currency,
				PretaxAmount:     costMap[tea.StringValue(item.BusinessCodeName)+tea.StringValue(item.PayModeName)],
			}
			if param.Granularity == types.Daily {
				temp.BillingDate = param.BillingDate
			}
			billingItem = append(billingItem, temp)
			checkMap[tea.StringValue(item.BusinessCodeName)+tea.StringValue(item.PayModeName)] = true
		}
		return billingItem, nil
	}
	for _, cost := range costMap {
		totalCost += cost
	}
	temp := types.AccountBillItem{
		Currency:     currency,
		PretaxAmount: totalCost,
	}
	if param.Granularity == types.Daily {
		temp.BillingDate = param.BillingDate
	}
	billingItem = append(billingItem, temp)
	return billingItem, nil
}

func convPipCode(bizCode *string) types.PipCode {
	return types.PipCode(*bizCode)
}

func convPretaxAmount(price *string) float64 {
	if price == nil {
		return 0
	}
	priceFloat, _ := strconv.ParseFloat(*price, 64)
	return priceFloat
}

func convPayTime2YM(payTime *string) (string, error) {
	if payTime == nil {
		return "", errors.New("PayTime empty")
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", *payTime, time.Local)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02"), nil
}
func parseDateStartEndTime(date string) (string, string, error) {
	if date == "" {
		return "", "", errors.New("date empty")
	}
	t, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return "", "", err
	}
	dateFormatted := t.Format("2006-01-02")

	return dateFormatted + " 00:00:00", dateFormatted + " 23:59:59", nil
}
func sumComponentSet(componentSet []*billing.BillDetailComponent) (result float64) {
	for _, v := range componentSet {
		result += convPretaxAmount(v.RealCost)
	}
	return
}
func convSubscriptionType(subscriptionType string) cloud.SubscriptionType {
	switch subscriptionType {
	case "包年包月":
		return cloud.PrePaid
	case "按量计费":
		return cloud.PostPaid
	}
	return "undefined"
}
func convCurrency(priceUnit string) (currency string) {
	strs := strings.Split(priceUnit, "/")
	if len(strs) == 0 {
		return
	}
	switch strs[0] {
	case "元":
		currency = "CNY"
	case "刀":
		currency = "USD"
	}
	return
}

// DescribeMetricList api doc https://cloud.tencent.com/document/api/248/31014
func (p *TencentCloud) DescribeMetricList(ctx context.Context, param types.DescribeMetricListRequest) (types.DescribeMetricList, error) {

	time.Sleep(500 * time.Millisecond)

	//check undefined type
	if param.MetricName == types.Undefined {
		return types.DescribeMetricList{}, errors.New("param 'metricName' undefined")
	}

	request := monitor.NewGetMonitorDataRequest()

	//time
	start := param.StartTime.Format(time.RFC3339)
	end := param.EndTime.Format(time.RFC3339)
	request.StartTime = &start
	request.EndTime = &end

	//period
	period, err := strconv.ParseUint(param.Period, 10, 64)
	if err != nil {
		return types.DescribeMetricList{}, err
	}
	request.Period = &period

	//instances
	var instances []*monitor.Instance
	var dimensions []*monitor.Dimension
	name := "InstanceId"
	for _, ins := range param.Filter.InstanceIds {

		dimensions = append(dimensions, &monitor.Dimension{
			Name:  &name,
			Value: &ins,
		})

	}
	instances = append(instances, &monitor.Instance{
		Dimensions: dimensions,
	})
	request.Instances = instances

	//namespace
	nameSpace := "QCE/CVM"
	request.Namespace = &nameSpace

	//metric name
	var metricName string
	switch param.MetricName {
	case types.MetricItemCPUUtilization:

		metricName = "CPUUsage"
		break
		//内存使用情况
	case types.MetricItemMemoryUsedUtilization:
		metricName = "MemUsage"
		break
	}
	request.MetricName = &metricName

	monitorData, err := p.monitorClient.GetMonitorData(request)
	if err != nil {
		return types.DescribeMetricList{}, err
	}
	fmt.Println(monitorData)

	//Timestamp         int64	时间戳
	//InstanceId        string	实例id
	//Min, Max, Average float64 最小，最大，平均值

	var metricSamples []types.MetricSample

	for _, d := range monitorData.Response.DataPoints {

		//组合时间戳 数据
		var min float64
		var max float64
		var ave float64
		for index, t := range d.Values {

			ave += *t
			if index == 0 {
				min = *t
				max = *t

				continue
			}

			if min > *t {
				min = *t
			}

			if max < *t {
				max = *t
			}

		}

		metricSamples = append(metricSamples, types.MetricSample{
			Timestamp:  param.StartTime.UnixMilli(),
			InstanceId: param.Filter.InstanceIds[0],
			Min:        min,
			Max:        max,
			Average:    ave / float64(len(d.Values)),
		})

	}

	return types.DescribeMetricList{
		List: metricSamples,
	}, nil
}

// DescribeRegions get all available regions of the current account
func (p *TencentCloud) DescribeRegions(ctx context.Context, param types.DescribeRegionsRequest) (types.DescribeRegions, error) {

	request := cvm.NewDescribeRegionsRequest()
	regions, err := p.cvmClient.DescribeRegions(request)

	if err != nil {
		return types.DescribeRegions{}, err
	}

	regionSet := regions.Response.RegionSet

	var ItemRegions []types.ItemRegion
	for _, item := range regionSet {
		itemRegion := types.ItemRegion{}
		itemRegion.RegionId = *item.Region
		itemRegion.LocalName = *item.RegionName

		ItemRegions = append(ItemRegions, itemRegion)
	}

	return types.DescribeRegions{
		List: ItemRegions,
	}, nil
}
func (p *TencentCloud) DescribeInstanceBill(ctx context.Context, param types.DescribeInstanceBillRequest, isAll bool) (types.DescribeInstanceBill, error) {
	return types.DescribeInstanceBill{}, nil
}

func (p *TencentCloud) QueryAvailableInstances(ctx context.Context, param types.QueryAvailableInstancesRequest) (types.QueryAvailableInstances, error) {
	return types.QueryAvailableInstances{}, nil
}

func formatChargeType(t string) cloud.SubscriptionType {

	switch t {
	case "PREPAID":
		return cloud.PrePaid
	case "POSTPAID_BY_HOUR":
		return cloud.PostPaid
	default:
		return cloud.Undefined
	}

}

// DescribeInstances get available instances of the current region

func (p *TencentCloud) DescribeInstances(_ context.Context, param types.DescribeInstancesRequest) (types.DescribeInstances, error) {

	request := cvm.NewDescribeInstancesRequest()

	var offset int64 = 0
	var limit int64 = 100
	var total int64 = 0
	request.Offset = &offset
	var instanceSet []*cvm.Instance
	for true {

		request.Offset = &offset
		request.Limit = &limit
		instancesRes, err := p.cvmClient.DescribeInstances(request)
		if err != nil {
			return types.DescribeInstances{}, err
		}

		instanceSet = append(instanceSet, instancesRes.Response.InstanceSet...)
		instanceSet = instancesRes.Response.InstanceSet
		total += *instancesRes.Response.TotalCount

		if int64(len(instancesRes.Response.InstanceSet)) < limit {
			break
		}

		offset += limit

	}

	var itemDescribeInstances []types.ItemDescribeInstance
	for _, instance := range instanceSet {
		var itemDescribeInstance = types.ItemDescribeInstance{
			InstanceId:         *instance.InstanceId,
			InstanceName:       *instance.InstanceName,
			RegionName:         *instance.Placement.Zone,
			SubscriptionType:   formatChargeType(*instance.InstanceChargeType),
			InternetChargeType: *instance.InstanceChargeType,
		}

		var publicIps []string
		for _, ip := range instance.PublicIpAddresses {
			publicIps = append(publicIps, *ip)
		}

		var privateIps []string
		for _, ip := range instance.PrivateIpAddresses {
			privateIps = append(privateIps, *ip)
		}

		itemDescribeInstance.PublicIpAddress = publicIps
		itemDescribeInstance.InnerIpAddress = privateIps

		itemDescribeInstances = append(itemDescribeInstances, itemDescribeInstance)
	}

	return types.DescribeInstances{TotalCount: int(total), List: itemDescribeInstances}, nil
}
