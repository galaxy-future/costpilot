package huawei

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	bss "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2/model"
	regionHuawei "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2/region"
)

var _AK = ""
var _SK = ""
var _REGION = "cn-north-1"

func TestHuaweiCloud_QueryAccountBill(t *testing.T) {
	type fields struct {
		bssClientOpt *bss.BssClient
	}
	type args struct {
		ctx   context.Context
		param types.QueryAccountBillRequest
	}

	auth := global.NewCredentialsBuilder().
		WithAk(_AK).
		WithSk(_SK).
		Build()

	bssClientOpt := bss.NewBssClient(bss.BssClientBuilder().WithRegion(regionHuawei.ValueOf(_REGION)).WithCredential(auth).Build())

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.DataInQueryAccountBill
		wantErr bool
	}{
		{
			name: "Monthly-Group_false",
			fields: fields{
				bssClientOpt: bssClientOpt,
			},
			args: args{
				ctx: context.Background(),
				param: types.QueryAccountBillRequest{
					BillingCycle: "2022-09",
					Granularity:  types.Monthly,
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Monthly-Group_true",
			fields: fields{
				bssClientOpt: bssClientOpt,
			},
			args: args{
				param: types.QueryAccountBillRequest{
					BillingCycle:     "2022-09",
					IsGroupByProduct: true,
					Granularity:      types.Monthly,
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Daily-Group_false",
			fields: fields{
				bssClientOpt: bssClientOpt,
			},
			args: args{
				param: types.QueryAccountBillRequest{
					BillingCycle: "2022-09",
					BillingDate:  "2022-09-06",
					Granularity:  types.Daily,
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Daily-Group_true",
			fields: fields{
				bssClientOpt: bssClientOpt,
			},
			args: args{
				param: types.QueryAccountBillRequest{
					BillingCycle:     "2022-09",
					BillingDate:      "2022-09-06",
					IsGroupByProduct: true,
					Granularity:      types.Daily,
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &HuaweiCloud{
				bssClientOpt: tt.fields.bssClientOpt,
			}
			got, err := p.QueryAccountBill(tt.args.ctx, tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryAccountBill() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("QueryAccountBill() got = %v, want %v", got, tt.want)
			t.Logf("QueryAccountBill() got = %+v", got)
		})
	}
}

func TestHuaweiCloud_QueryAccountBillResult(t *testing.T) {
	billItems := make([]types.AccountBillItem, 0)

	var data = []byte("{\"fee_records\":[{\"bill_date\":\"2020-12-06\",\"bill_type\":1,\"customer_id\":\"52190d93cb844a249c70fd1e1d416f8b\",\"region\":\"cn-north-1\",\"region_name\":\"CN North-Beijing1\",\"cloud_service_type\":\"hws.service.type.vpc\",\"resource_type\":\"hws.resource.type.pm\",\"effective_time\":\"2020-12-06T11:06:55Z\",\"expire_time\":\"2020-12-07T11:06:55Z\",\"resource_id\":\"4251f987c09c4d97a6b4784e4661f8ce\",\"resource_name\":\"hws.service.type.vpcname\",\"resource_tag\":\"resourceTag\",\"product_id\":\"90301-686010-0--0\",\"product_name\":\"全动态BGP弹性IP_包月_北京一 北京四\",\"product_spec_desc\":\"动态BGP弹性IP\",\"sku_code\":\"5_bgp\",\"spec_size\":40,\"spec_size_measure_id\":0,\"trade_id\":\"BC0883684711\",\"trade_time\":\"2020-12-06T11:07:00Z\",\"enterprise_project_id\":\"0\",\"enterprise_project_name\":\"default\",\"charge_mode\":\"1\",\"order_id\":\"CS21100100328BXN3\",\"period_type\":\"20\",\"usage_type\":\"sdjhgkf\",\"usage\":101,\"usage_measure_id\":1,\"free_resource_usage\":123,\"free_resource_measure_id\":1,\"ri_usage\":30,\"ri_usage_measure_id\":0,\"unit_price\":0,\"unit\":\"元/1个(次)\",\"official_amount\":34.96,\"discount_amount\":0.002,\"amount\":34.96,\"cash_amount\":1.23,\"credit_amount\":1.24,\"coupon_amount\":0.33,\"flexipurchase_coupon_amount\":22.5,\"stored_card_amount\":12.13,\"bonus_amount\":2.4,\"debt_amount\":-4.87,\"adjustment_amount\":2.58,\"measure_id\":1},{\"bill_date\":\"2020-12-05\",\"bill_type\":1,\"customer_id\":\"52190d93cb844a249c70fd1e1d416f8b\",\"region\":\"cn-north-1\",\"region_name\":\"CN North-Beijing1\",\"cloud_service_type\":\"hws.service.type.vpc\",\"resource_type\":\"hws.resource.type.ip\",\"effective_time\":\"2020-12-05T11:06:55Z\",\"expire_time\":\"2020-12-06T11:06:55Z\",\"resource_id\":\"4251f987c09c4d97a6b4784e4661f8ce\",\"resource_name\":\"hws.service.type.vpcname\",\"resource_tag\":\"resourceTag\",\"product_id\":\"00301-110660-0--0\",\"product_name\":\"调试15_4核8G_linux 包年\",\"product_spec_desc\":\"调试15_4核8G_linux\",\"sku_code\":\"comtest15.linux\",\"spec_size\":40,\"spec_size_measure_id\":0,\"trade_id\":\"BC0883684711\",\"trade_time\":\"2020-12-05T11:07:00Z\",\"enterprise_project_id\":\"0\",\"enterprise_project_name\":\"default\",\"charge_mode\":\"1\",\"order_id\":\"BC0883684711\",\"period_type\":\"20\",\"usage_type\":\"dsfhjgbk\",\"usage\":147,\"usage_measure_id\":1,\"free_resource_usage\":258,\"free_resource_measure_id\":1,\"ri_usage\":30,\"ri_usage_measure_id\":0,\"unit_price\":0,\"unit\":\"元/1个(次)\",\"official_amount\":0.81,\"discount_amount\":0.01,\"amount\":0.81,\"cash_amount\":2.25,\"credit_amount\":1.23,\"coupon_amount\":0.07,\"flexipurchase_coupon_amount\":0.4,\"stored_card_amount\":0.34,\"bonus_amount\":4.63,\"debt_amount\":-8.11,\"adjustment_amount\":3.69,\"measure_id\":1}],\"total_count\":2,\"currency\":\"CNY\"}")

	response := new(model.ListCustomerselfResourceRecordsResponse)

	if err := json.Unmarshal(data, &response); err != nil {
		t.Logf("QueryAccountBill() err = %+v", err)
	}
	fmt.Println(response)

	totalCount := response.TotalCount
	if len(billItems) == 0 {
		billItems = make([]types.AccountBillItem, 0, *totalCount)
	}
	billItems = append(billItems, testConvQueryAccountBill(response, *response.Currency)...)

	result := types.DataInQueryAccountBill{
		BillingCycle: *response.Currency,
		AccountID:    "",
		TotalCount:   len(billItems),
		AccountName:  "",
		Items: types.ItemsInQueryAccountBill{
			Item: billItems,
		},
	}

	fmt.Println(result)
}

// convQueryAccountBill
func testConvQueryAccountBill(response *model.ListCustomerselfResourceRecordsResponse, currency string) []types.AccountBillItem {
	if response == nil {
		return nil
	}

	feeRecords := *response.FeeRecords
	result := make([]types.AccountBillItem, 0, len(feeRecords))
	for _, v := range feeRecords {
		//standardPipCode := convPipCode(*v.CloudServiceTypeName)
		standardPipCode := types.ECS
		item := types.AccountBillItem{
			PipCode:          standardPipCode,
			ProductName:      *v.ProductName,
			BillingDate:      *v.BillDate,                              // has date when Granularity=DAILY
			SubscriptionType: testConvSubscriptionType("Subscription"), // 先写死
			Currency:         currency,
			PretaxAmount:     *v.Amount,
		}
		result = append(result, item)
	}

	return result
}

func testConvPipCode(pipCode string) types.PipCode {
	//switch pipCode {
	//case "oss":
	//	return types.S3
	//}
	// 暂不启用转换，直接返回
	return types.PipCode(pipCode)
}

func testConvSubscriptionType(subscriptionType string) cloud.SubscriptionType {
	switch subscriptionType {
	case "Subscription":
		return cloud.PrePaid
	case "PayAsYouGo":
		return cloud.PostPaid
	}
	return "undefined"
}
