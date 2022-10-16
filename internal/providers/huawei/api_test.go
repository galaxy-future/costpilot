package huawei

import (
	"context"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	bss "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2"
	regionHuawei "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2/region"
	"testing"

	"github.com/galayx-future/costpilot/internal/providers/types"
)

var _AK = "WL1NEDSGD9M0H1VYX1C8"
var _SK = "90M1drAdT7NLnPAX7MvKpDEh84VzLCQvKaXFPCtx"
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
