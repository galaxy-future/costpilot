package huawei

import (
	"context"
	"github.com/galayx-future/costpilot/internal/constants/cloud"
	"github.com/galayx-future/costpilot/internal/providers/types"
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

func (*HuaweiCloud) ProviderType() string {
	return cloud.HuaweiCloud
}

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

func convQueryAccountBill(response *model.ListCustomerselfResourceRecordsResponse, currency string) []types.AccountBillItem {
	if response == nil {
		return nil
	}

	feeRecords := *response.FeeRecords
	result := make([]types.AccountBillItem, 0, len(feeRecords))
	for _, v := range feeRecords {
		standardPipCode := convPipCode(*v.CloudServiceType)
		item := types.AccountBillItem{
			PipCode:          standardPipCode,
			ProductName:      *v.ProductName,
			BillingDate:      *v.BillDate, // has date when Granularity=DAILY
			SubscriptionType: convSubscriptionType(*v.ChargeMode),
			Currency:         currency,
			PretaxAmount:     *v.Amount,
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
