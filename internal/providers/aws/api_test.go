package aws

import (
	"context"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"

	"github.com/galaxy-future/costpilot/internal/providers/types"
)

var (
	_AK = ""
	_SK = ""
	cli *AWSCloud
)

func init() {
	c, err := New(_AK, _SK, "ap-northeast-1")
	if err != nil {
		return
	}
	cli = c
}

func TestAWSCloud_QueryAccountBill(t *testing.T) {
	type args struct {
		ctx   context.Context
		param types.QueryAccountBillRequest
	}
	awsCloud, err := New(_AK, _SK, "ap-northeast-1")
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

func TestAWSCloud_DescribeRegions(t *testing.T) {
	regions, err := cli.DescribeRegions(context.Background(), types.DescribeRegionsRequest{})
	t.Log(regions)
	t.Log(err)
	assert.Equal(t, len(regions.List), 17)
}

func TestAWSCloud_DescribeInstances(t *testing.T) {
	instances, err := cli.DescribeInstances(context.Background(), types.DescribeInstancesRequest{
		//InstanceIds: []string{"i-09491347f48116001"},
	})
	t.Log(instances)
	t.Log(err)
}

func TestAWSCloud_DescribeMetricList(t *testing.T) {
	startTime, _ := time.ParseInLocation("2006-01-02", "2022-12-07", time.Local)
	endTime := startTime.AddDate(0, 0, +1)
	describeMetricList, err := cli.DescribeMetricList(context.Background(), types.DescribeMetricListRequest{
		StartTime: startTime,
		EndTime:   endTime,
		Period:    "86400",
		Filter: types.MetricListInstanceFilter{
			InstanceIds: []string{"i-09491347f48116001"},
		},
		MetricName: types.MetricItemCPUUtilization,
	})
	t.Log(describeMetricList)
	t.Log(err)
	describeMetricList, err = cli.DescribeMetricList(context.Background(), types.DescribeMetricListRequest{
		StartTime: startTime,
		EndTime:   endTime,
		Period:    "86400",
		Filter: types.MetricListInstanceFilter{
			InstanceIds: []string{"i-09491347f48116001"},
		},
		MetricName: types.MetricItemMemoryUsedUtilization,
	})
	t.Log(describeMetricList)
	t.Log(err)
}
