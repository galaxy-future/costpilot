package tencent

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	billing "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/billing/v20180709"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"github.com/galaxy-future/costpilot/internal/providers/types"
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

var (
	_AK = ""
	_SK = ""
	// _regionId = "ap-guangzhou"
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
	billingClient, err := billing.NewClient(credential, "", cpf)
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
					BillingCycle: "2022-01",
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

func Test_parseDateStartEndTime(t *testing.T) {
	startTime, endTime, err := parseDateStartEndTime("2022-09-09")
	assert.Nil(t, err)
	assert.Equal(t, "2022-09-09 00:00:00", startTime)
	assert.Equal(t, "2022-09-09 23:59:59", endTime)
}

/*func Test_convQueryAccountBill1(t *testing.T) {
	billListJson := `[{"ActionType":"prepay_renew","ActionTypeName":"包年包月续费","BillId":"20201102400000425173641","BusinessCode":"p_cvm","BusinessCodeName":"云服务器CVM","ComponentSet":[{"CashPayAmount":"17.46","ComponentCode":"v_cvm_bandwidth","ComponentCodeName":"带宽","ContractPrice":"17.46","Cost":"18","Discount":"0.97","IncentivePayAmount":"0","ItemCode":"sv_cvm_bandwidth_prepay","ItemCodeName":"带宽-按带宽计费","PriceUnit":"元/Mbps/月","RealCost":"17.46","ReduceType":"折扣","SinglePrice":"18","SpecifiedPrice":"18","TimeSpan":"1","TimeUnitName":"月","UsedAmount":"1","UsedAmountUnit":"Mbps","VoucherPayAmount":"0"},{"CashPayAmount":"17.46","ComponentCode":"virtual_v_cvm_compute","ComponentCodeName":"运算组件","ContractPrice":"17.46","Cost":"18","Discount":"0.97","IncentivePayAmount":"0","ItemCode":"virtual_v_cvm_compute_sa2","ItemCodeName":"运算组件-标准型SA2-1核1G","PriceUnit":"元/个/月","RealCost":"17.46","ReduceType":"折扣","SinglePrice":"18","SpecifiedPrice":"18","TimeSpan":"1","TimeUnitName":"月","UsedAmount":"1","UsedAmountUnit":"个","VoucherPayAmount":"0"},{"CashPayAmount":"16.98","ComponentCode":"v_cvm_rootdisk","ComponentCodeName":"系统盘","ContractPrice":"0.3395","Cost":"17.5","Discount":"0.97","IncentivePayAmount":"0","ItemCode":"sv_cvm_rootdisk_cbspremium","ItemCodeName":"高效云系统盘","PriceUnit":"元/GB/月","RealCost":"16.98","ReduceType":"折扣","SinglePrice":"0.35","SpecifiedPrice":"0.35","TimeSpan":"1","TimeUnitName":"月","UsedAmount":"50","UsedAmountUnit":"GB","VoucherPayAmount":"0"}],"FeeBeginTime":"2020-11-02 12:05:15","FeeEndTime":"2020-12-02 12:05:15","OperateUin":"909619400","OrderId":"20201102400000425173641","OwnerUin":"909619400","PayModeName":"包年包月","PayTime":"2020-11-02 02:29:57","PayerUin":"909619400","ProductCode":"sp_cvm_sa2","ProductCodeName":"云服务器CVM-标准型SA2","ProjectId":"0","ProjectName":"默认项目","RegionId":"16","RegionName":"西南地区（成都）","ResourceId":"ins-m1okcccv","ResourceName":"windows-1GB-cd-1880","Tags":null,"ZoneName":"成都一区"}]`
	var billList []*billing.BillDetail
	_ = json.Unmarshal([]byte(billListJson), &billList)
	itemList1, err1 := convQueryAccountBill(false, billList)
	assert.Nil(t, err1)
	assert.Equal(t, 3, len(itemList1))

	itemList2, err2 := convQueryAccountBill(true, billList)
	assert.Nil(t, err2)
	if assert.Equal(t, 1, len(itemList2)) {
		assert.Equal(t, 17.46+17.46+16.98, itemList2[0].PretaxAmount)
	}
}*/

func Test_convPretaxAmount1(t *testing.T) {
	price1 := "12.35677"
	type args struct {
		price *string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "normal",
			args: args{price: &price1},
			want: 12.35,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, convPretaxAmount(tt.args.price), "convPretaxAmount(%v)", tt.args.price)
		})
	}
}

var p *TencentCloud

func TestMain(m *testing.M) {
	var err error
	p, err = New(_AK, _SK, "")
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
func TestQueryAccountBill(t *testing.T) {
	tests := []types.QueryAccountBillRequest{
		{
			BillingCycle:     "2022-01",
			BillingDate:      "",
			IsGroupByProduct: true,
			Granularity:      "MONTHLY",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			resp, err := p.QueryAccountBill(context.Background(), tt)
			if err != nil {
				t.Errorf("%#v", err)
			}
			fmt.Println(resp)
		})
	}

}
