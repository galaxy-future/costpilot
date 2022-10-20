package huawei

import (
	"context"
	"fmt"
	"github.com/galayx-future/costpilot/internal/constants/cloud"
	"github.com/galayx-future/costpilot/internal/providers/types"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	bss "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2/model"
	regionHuawei "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2/region"
	"time"
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
func (*HuaweiCloud) ProviderType() string {
	return cloud.HuaweiCloud
}

// QueryAccountBill
func (p *HuaweiCloud) QueryAccountBill(ctx context.Context, param types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error) {
	var err error
	billItems := make([]types.AccountBillItem, 0)
	request := &model.ListCustomerselfResourceRecordsRequest{}
	request.Cycle = param.BillingCycle
	offsetRequest := int32(0)
	request.Offset = &offsetRequest
	limitRequest := int32(10)
	request.Limit = &limitRequest

	pageNum := int32(0)
	response := new(model.ListCustomerselfResourceRecordsResponse)
	for {
		response, err = p.bssClientOpt.ListCustomerselfResourceRecords(request)

		if err != nil {
			return types.DataInQueryAccountBill{}, err
		}
		totalCount := response.TotalCount
		if len(billItems) == 0 {
			billItems = make([]types.AccountBillItem, 0, *totalCount)
		}
		billItems = append(billItems, convQueryAccountBill(response, *response.Currency)...)
		if len(billItems) >= int(*totalCount) {
			break
		}
		pageNum++
		request.Offset = &pageNum
	}
	result := types.DataInQueryAccountBill{
		BillingCycle: *response.Currency,
		AccountID:    "",
		TotalCount:   len(billItems),
		AccountName:  "",
		Items: types.ItemsInQueryAccountBill{
			Item: billItems,
		},
	}

	return result, nil
}

// convQueryAccountBill
func convQueryAccountBill(response *model.ListCustomerselfResourceRecordsResponse, currency string) []types.AccountBillItem {
	if response == nil {
		return nil
	}

	feeRecords := *response.FeeRecords
	result := make([]types.AccountBillItem, 0, len(feeRecords))
	for _, v := range feeRecords {
		// 类型这里需要匹配，在寻找接口
		//standardPipCode := convPipCode(*v.CloudServiceTypeName)
		// todo 查询一下华为的资源类型
		standardPipCode := types.ECS
		item := types.AccountBillItem{
			PipCode:          standardPipCode,
			ProductName:      convProductName(standardPipCode, *v.ProductName),
			BillingDate:      *v.BillDate,                          // has date when Granularity=DAILY
			SubscriptionType: convSubscriptionType("Subscription"), // todo 先写死
			Currency:         currency,
			PretaxAmount:     *v.Amount,
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

func (p *HuaweiCloud) ListServiceTypes() (*model.ListServiceTypesResponse, error) {
	var err error
	billItems := make([]types.ServiceType, 0)

	request := &model.ListServiceTypesRequest{}
	offsetRequest := int32(0)
	request.Offset = &offsetRequest
	limitRequest := int32(10)
	request.Limit = &limitRequest

	pageNum := int32(0)
	response := &model.ListServiceTypesResponse{}
	for {
		response, err := p.bssClientOpt.ListServiceTypes(request)

		fmt.Println("request:{}, response:{}", request, response)
		if err == nil {
			fmt.Printf("%+v\n", response)
		} else {
			fmt.Println(err)
		}

		if err != nil {
			return nil, err
		}
		totalCount := response.TotalCount
		if len(billItems) == 0 {
			billItems = make([]types.ServiceType, 0, *totalCount)
		}

		if len(billItems) >= int(*totalCount) {
			break
		}

		for _, v := range *response.ServiceTypes {
			serviceType := types.ServiceType{}
			serviceType.ServiceTypeName = *v.ServiceTypeName
			serviceType.ServiceTypeCode = *v.ServiceTypeCode
			serviceType.Abbreviation = *v.Abbreviation
			billItems = append(billItems, serviceType)
		}

		pageNum++
		request.Offset = &pageNum
		time.Sleep(2 * time.Second)
		fmt.Println("request:{}", request)
	}

	fmt.Println("billItems:{}", billItems)

	return response, err
}
