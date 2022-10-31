package tencent

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/galayx-future/costpilot/internal/constants/cloud"
	"github.com/galayx-future/costpilot/internal/providers/types"
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
	billingClient, err := billing.NewClient(credential, regionId, cpf)
	if err != nil {
		return nil, err
	}
	return &TencentCloud{billingClient: billingClient}, nil
}

// ProviderType
func (*TencentCloud) ProviderType() string {
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

	itemList, err := convQueryAccountBill(param.IsGroupByProduct, allBillList)
	if err != nil {
		return types.DataInQueryAccountBill{}, err
	}
	return types.DataInQueryAccountBill{
		BillingCycle: param.BillingCycle,
		TotalCount:   len(itemList),
		Items:        types.ItemsInQueryAccountBill{Item: itemList},
	}, nil
}

func convQueryAccountBill(isGroupByProduct bool, billList []*billing.BillDetail) ([]types.AccountBillItem, error) {
	var billingItem []types.AccountBillItem
	for _, item := range billList {
		payTime, err := convPayTime2YM(item.PayTime)
		if err != nil {
			return nil, err
		}

		if isGroupByProduct {
			// 各组件金额相加
			var amount float64 = 0
			for _, i := range item.ComponentSet {
				amount += convPretaxAmount(i.RealCost)
			}
			billingItem = append(billingItem, types.AccountBillItem{
				PipCode:      convPipCode(item.BusinessCode),
				ProductName:  *item.BusinessCodeName,
				BillingDate:  payTime,
				PretaxAmount: amount,
			})
		} else {
			for _, i := range item.ComponentSet {
				productName := fmt.Sprintf("%s-%s", *item.BusinessCodeName, *i.ItemCodeName)
				billingItem = append(billingItem, types.AccountBillItem{
					PipCode:      convPipCode(item.BusinessCode),
					ProductName:  productName,
					BillingDate:  payTime,
					PretaxAmount: convPretaxAmount(i.RealCost),
				})
			}
		}
	}
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
	fmt.Println(priceFloat)
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
