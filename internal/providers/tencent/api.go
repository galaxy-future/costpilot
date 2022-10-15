package tencent

import (
	"context"
	"strconv"

	"github.com/galayx-future/costpilot/internal/constants/cloud"
	"github.com/galayx-future/costpilot/internal/providers/types"
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
	request := billing.NewDescribeBillSummaryByProductRequest()
	request.BeginTime = &param.BillingCycle
	request.EndTime = &param.BillingCycle
	response, err := p.billingClient.DescribeBillSummaryByProduct(request)
	if err != nil {
		return types.DataInQueryAccountBill{}, err
	}

	return types.DataInQueryAccountBill{
		BillingCycle: param.BillingCycle,
		TotalCount:   len(response.Response.SummaryOverview),
		Items:        types.ItemsInQueryAccountBill{Item: convQueryAccountBill(response.Response)},
	}, nil
}

func convQueryAccountBill(responseParams *billing.DescribeBillSummaryByProductResponseParams) []types.AccountBillItem {
	var billingItem []types.AccountBillItem
	for _, item := range responseParams.SummaryOverview {
		billingItem = append(billingItem, types.AccountBillItem{
			PipCode:      convPipCode(item.BusinessCode),
			ProductName:  *item.BusinessCodeName,
			BillingDate:  *item.BillMonth,
			PretaxAmount: convPretaxAmount(item.RealTotalCost),
		})
	}
	return billingItem
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
