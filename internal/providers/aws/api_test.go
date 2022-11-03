package aws

import (
	"context"
	"reflect"
	"testing"

	"github.com/galaxy-future/costpilot/internal/providers/types"
)

var _AK = ""
var _SK = ""

func TestAWSCloud_QueryAccountBill(t *testing.T) {
	type args struct {
		ctx   context.Context
		param types.QueryAccountBillRequest
	}
	awsCloud, err := New(_AK, _SK, "cn-north-1")
	if err != nil {
		t.Errorf("AWSCloud.New error=%v", err)
	}
	tests := []struct {
		name    string
		p       *AWSCloud
		args    args
		want    *types.DataInQueryAccountBill
		wantErr bool
	}{
		{
			name: "Monthly-Group_false",
			p:    awsCloud,
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
			p:    awsCloud,
			args: args{
				ctx: context.Background(),
				param: types.QueryAccountBillRequest{
					BillingCycle:     "2022-07",
					IsGroupByProduct: true,
					Granularity:      types.Monthly,
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Daily-Group_false",
			p:    awsCloud,
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
			p:    awsCloud,
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
			got, err := tt.p.QueryAccountBill(tt.args.ctx, tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("AWSCloud.QueryAccountBill() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AWSCloud.QueryAccountBill() = %v, want %v", got, tt.want)
			}
			t.Logf("QueryAccountBill() got = %+v", got)
		})
	}
}
