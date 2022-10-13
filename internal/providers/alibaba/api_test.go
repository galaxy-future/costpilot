package alibaba

import (
	"context"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"github.com/galayx-future/costpilot/internal/providers/types"
)

var _AK = "ak_test_123"
var _SK = "sk_test_123"

func TestAlibabaCloud_QueryAccountBill(t *testing.T) {
	type fields struct {
		bssClientOpt *bssopenapi.Client
	}
	type args struct {
		ctx   context.Context
		param types.QueryAccountBillRequest
	}
	bssClientOpt, err := bssopenapi.NewClientWithOptions("cn-beijing", sdk.NewConfig(), credentials.NewAccessKeyCredential(_AK, _SK))
	if err != nil {
		t.Fatal(err)
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
			p := &AlibabaCloud{
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
