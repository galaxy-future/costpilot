package tencent

import (
	"context"
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
)

type TencentCloud struct {
	billingClient *billing.Client
}

func New(ak, sk, regionId string) (*TencentCloud, error) {
	credential := common.NewCredential(ak, sk)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = _billingEndpoint
	billingClient, err := billing.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}
	return &TencentCloud{billingClient: billingClient}, nil
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

func (p *TencentCloud) DescribeMetricList(ctx context.Context, param types.DescribeMetricListRequest) (types.DescribeMetricList, error) {
	// TODO implement me
	panic("implement me")
}

func (p *TencentCloud) DescribeInstanceAttribute(ctx context.Context, param types.DescribeInstanceAttributeRequest) (types.DescribeInstanceAttribute, error) {
	// TODO implement me
	panic("implement me")
}

func (p *TencentCloud) DescribeRegions(ctx context.Context, param types.DescribeRegionsRequest) (types.DescribeRegions, error) {
	// TODO implement me
	panic("implement me")
}

func (p *TencentCloud) DescribeInstanceBill(ctx context.Context, param types.DescribeInstanceBillRequest, isAll bool) (types.DescribeInstanceBill, error) {
	// TODO implement me
	panic("implement me")
}

func (p *TencentCloud) QueryAvailableInstances(ctx context.Context, param types.QueryAvailableInstancesRequest) (types.QueryAvailableInstances, error) {
	// TODO implement me
	panic("implement me")
}
