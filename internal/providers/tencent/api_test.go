package tencent

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/galayx-future/costpilot/internal/providers/types"
	billing "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/billing/v20180709"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

func Test_convPretaxAmount(t *testing.T) {
	type args struct {
		price *string
	}
	price1 := "13.34"
	float1 := 13.34
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "valid",
			args: args{price: &price1},
			want: float1,
		},
		{
			name: "null",
			args: args{price: nil},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convPretaxAmount(tt.args.price); got != tt.want {
				t.Errorf("convPretaxAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convQueryAccountBill(t *testing.T) {
	mockResponseJson := `{"Response":{"Ready":1,"SummaryTotal":{"RealTotalCost":"1458.00000000","TotalCost":"1458.00000000","VoucherPayAmount":"0.00000000","IncentivePayAmount":"0.00000000","CashPayAmount":"1458.00000000","TransferPayAmount":"0.00000000"},"SummaryOverview":[{"BusinessCode":"p_ssl","RealTotalCost":"1458.00000000","TotalCost":"1458.00000000","CashPayAmount":"1458.00000000","IncentivePayAmount":"0.00000000","VoucherPayAmount":"0.00000000","TransferPayAmount":"0.00000000","RealTotalCostRatio":"100.00","BillMonth":"2022-07","BusinessCodeName":"SSL证书"}],"RequestId":"67cd3369-b022-4a6a-818e-7ba5a05cb5d7"}}`
	type args struct {
		response *billing.DescribeBillSummaryByProductResponse
	}
	var mockData billing.DescribeBillSummaryByProductResponse
	err := json.Unmarshal([]byte(mockResponseJson), &mockData)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	t.Log(mockData)
	tests := []struct {
		name    string
		args    args
		itemNum int
	}{
		{
			name:    "item num = 1",
			args:    args{response: &mockData},
			itemNum: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convQueryAccountBill(tt.args.response.Response)
			if !reflect.DeepEqual(len(got), tt.itemNum) {
				t.Errorf("convQueryAccountBill() = %v, want %v", len(got), tt.itemNum)
			}
		})
	}
}

var (
	_AK       = "ak_test_123"
	_SK       = "sk_test_123"
	_regionId = "ap-guangzhou"
)

func TestTencentCloud_QueryAccountBill(t *testing.T) {
	type fields struct {
		billingClient *billing.Client
	}
	type args struct {
		ctx   context.Context
		param types.QueryAccountBillRequest
	}
	credential := common.NewCredential(_AK, _SK)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = _billingEndpoint
	billingClient, err := billing.NewClient(credential, _regionId, cpf)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
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
				billingClient: billingClient,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &TencentCloud{
				billingClient: tt.fields.billingClient,
			}
			got, err := p.QueryAccountBill(tt.args.ctx, tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryAccountBill() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("QueryAccountBill() got = %+v", got)
		})
	}
}
