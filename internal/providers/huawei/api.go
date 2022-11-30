package huawei

import (
	"context"
	"strconv"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/galaxy-future/costpilot/tools/limiter"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	bss "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2/model"
	regionHuawei "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2/region"
)

type HuaweiCloud struct {
	bssClientOpt *bss.BssClient
}

func New(AK, SK, region string) (*HuaweiCloud, error) {
	auth := global.NewCredentialsBuilder().
		WithAk(AK).
		WithSk(SK).
		Build()

	bssClientOpt := bss.NewBssClient(bss.BssClientBuilder().WithRegion(regionHuawei.ValueOf(region)).WithCredential(auth).Build())

	return &HuaweiCloud{
		bssClientOpt: bssClientOpt,
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
func convQueryAccountBillByMonth(param types.QueryAccountBillRequest, response *model.ShowCustomerMonthlySumResponse) []types.AccountBillItem {
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
func convQueryAccountBill(response *model.ListCustomerselfResourceRecordsResponse) []types.AccountBillItem {
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
	//1:包年/包月
	case "1":
		return cloud.PrePaid
	//3：按需
	case "3":
		return cloud.PostPaid
	}
	return "undefined"
}

func convPipCode(pipCode string) types.PipCode {
	switch pipCode {
	//弹性云服务器
	case "hws.service.type.ec2":
		return types.ECS
	//弹性公网IP
	case "hws.service.type.eip":
		return types.EIP
	//对象存储服务
	case "hws.service.type.obs":
		return types.S3
		//NAT网关
	case "hws.service.type.natgateway":
		return types.NAT
		//弹性文件服务
	case "hws.service.type.sfs":
		return types.NAS
		//弹性负载均衡
	case "hws.service.type.elb":
		return types.SLB
		//分布式缓存服务
	case "hws.service.type.dcs":
		return types.KVSTORE
		//云桌面
	case "hws.service.type.vdi":
		return types.GWS
		// 企业网络部署规划设计服务
	case "hws.resource.type.pds.en":
		return types.CBN
	}
	return types.PipCode(pipCode)
}

func (p *HuaweiCloud) DescribeMetricList(ctx context.Context, param types.DescribeMetricListRequest) (types.DescribeMetricList, error) {
	return types.DescribeMetricList{}, nil
}

func (p *HuaweiCloud) DescribeInstanceAttribute(ctx context.Context, param types.DescribeInstanceAttributeRequest) (types.DescribeInstanceAttribute, error) {
	// TODO implement me
	return types.DescribeInstanceAttribute{}, nil
}

func (p *HuaweiCloud) DescribeRegions(ctx context.Context, param types.DescribeRegionsRequest) (types.DescribeRegions, error) {
	return types.DescribeRegions{}, nil
}

func (p *HuaweiCloud) DescribeInstanceBill(ctx context.Context, param types.DescribeInstanceBillRequest, isAll bool) (types.DescribeInstanceBill, error) {
	// TODO implement me
	return types.DescribeInstanceBill{}, nil
}

func (p *HuaweiCloud) QueryAvailableInstances(ctx context.Context, param types.QueryAvailableInstancesRequest) (types.QueryAvailableInstances, error) {
	// TODO implement me
	return types.QueryAvailableInstances{}, nil
}

func (p *HuaweiCloud) queryAccountBillByMonth(ctx context.Context, param types.QueryAccountBillRequest) (result types.DataInQueryAccountBill, err error) {

	billItems := make([]types.AccountBillItem, 0)
	request := &model.ShowCustomerMonthlySumRequest{}
	request.BillCycle = param.BillingCycle
	pageNum := int32(0)
	request.Offset = tea.Int32(0)
	request.Limit = tea.Int32(10)
	response := new(model.ShowCustomerMonthlySumResponse)
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
	request := &model.ListCustomerselfResourceRecordsRequest{}
	request.Cycle = param.BillingCycle
	request.BillDateBegin = tea.String(param.BillingDate)
	request.BillDateEnd = tea.String(param.BillingDate)
	includeZeroRecordRequest := false
	request.IncludeZeroRecord = &includeZeroRecordRequest
	request.Offset = tea.Int32(0)
	request.Limit = tea.Int32(10)
	pageNum := int32(0)
	response := new(model.ListCustomerselfResourceRecordsResponse)
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
