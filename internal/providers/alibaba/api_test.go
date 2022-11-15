package alibaba

import (
	"context"
	"testing"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"github.com/galaxy-future/costpilot/internal/providers/types"
)

var (
	_AK = "ak_test_123"
	_SK = "sk_test_123"

	cli *AlibabaCloud
)

func init() {
	c, err := New(_AK, _SK, "cn-hangzhou")
	if err != nil {
		return
	}
	cli = c
}

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
			// if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("QueryAccountBill() got = %v, want %v", got, tt.want)
			t.Logf("QueryAccountBill() got = %+v", got)
		})
	}
}

func TestAlibabaCloud_DescribeMetricList(t *testing.T) {
	p, err := New(_AK, _SK, "cn-shenzhen")
	if err != nil {
		return
	}
	startTime, _ := time.Parse("2006-01-02", "2022-11-10")
	endTime, _ := time.Parse("2006-01-02", "2022-11-11")
	got, err := p.DescribeMetricList(nil, types.DescribeMetricListRequest{
		MetricName: types.MetricItemMemoryUsedUtilization,
		Period:     "86400",
		StartTime:  startTime,
		EndTime:    endTime,
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(got)
}

func TestAlibabaCloud_DescribeRegions(t *testing.T) {
	got, err := cli.DescribeRegions(nil, types.DescribeRegionsRequest{
		ResourceType: types.ResourceTypeInstance,
		Language:     types.RegionLanguageZHCN,
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(got)
}

func TestAlibabaCloud_DescribeInstanceAttribute(t *testing.T) {
	got, err := cli.DescribeInstanceAttribute(nil, types.DescribeInstanceAttributeRequest{
		InstanceId: "i-wz9ctvduhhj02x4nc5k7",
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(got)
}

func TestAlibabaCloud_DescribeInstanceBill(t *testing.T) {
	rsp, err := cli.DescribeInstanceBill(context.TODO(), types.DescribeInstanceBillRequest{
		BillingCycle: "2022-11",
		Granularity:  types.Monthly,
		InstanceId:   "i-wz95ivyghpphzwqls6mq",
	}, true)
	t.Log(rsp, err)
}

func TestAlibabaCloud_QueryAvailableInstances(t *testing.T) {
	_, err := cli.QueryAvailableInstances(context.TODO(), types.QueryAvailableInstancesRequest{
		InstanceIdList: []string{"i-wz9g67k0g3582e1z8j60"},
	})
	t.Log(err)
}
