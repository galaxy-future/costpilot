package aws

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/galaxy-future/costpilot/internal/providers/types"
)

func init() {
	c, err := New(_AK, _SK, "ap-northeast-1")
	if err != nil {
		return
	}
	cli = c
}
func TestAWSCloud_DescribeRegions(t *testing.T) {
	regions, err := cli.DescribeRegions(context.Background(), types.DescribeRegionsRequest{})
	t.Log(regions)
	t.Log(err)
	assert.Equal(t, len(regions.List), 17)
}

func TestAWSCloud_DescribeInstances(t *testing.T) {
	instances, err := cli.DescribeInstances(context.Background(), types.DescribeInstancesRequest{
		InstanceIds: []string{"i-09491347f48116001"},
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
