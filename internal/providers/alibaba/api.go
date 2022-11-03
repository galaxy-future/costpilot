package alibaba

import (
	"context"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
)

type AlibabaCloud struct {
	bssClientOpt *bssopenapi.Client
}

func New(AK, SK, region string) (*AlibabaCloud, error) {
	bssClientOpt, err := bssopenapi.NewClientWithOptions(region, sdk.NewConfig().WithTimeout(10*time.Second), credentials.NewAccessKeyCredential(AK, SK))
	if err != nil {
		return nil, err
	}

	return &AlibabaCloud{
		bssClientOpt: bssClientOpt,
	}, nil
}

// ProviderType
func (*AlibabaCloud) ProviderType() string {
	return cloud.AlibabaCloud.String()
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
	case "Subscription":
		return cloud.PrePaid
	case "PayAsYouGo":
		return cloud.PostPaid
	}
	return "undefined"
}

func convPipCode(pipCode string) types.PipCode {
	//switch pipCode {
	//case "oss":
	//	return types.S3
	//}
	// 暂不启用转换，直接返回
	return types.PipCode(pipCode)
}

func convProductName(pipCode types.PipCode, defaultName ...string) string {
	//name := types.PidCode2Name(pipCode)
	//if name == types.Undefined && len(defaultName) != 0 {
	//	return defaultName[0]
	//}
	// 暂不启用转换，直接返回
	return defaultName[0]
}
