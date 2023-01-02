package baidu

import (
	"context"
	"testing"
	"time"

	"github.com/galaxy-future/costpilot/internal/providers/types"
)

var (
	_AK = "ak-test"
	_SK = "sk-test"

	cli *BaiduCloud
)

func init() {
	c, err := New(_AK, _SK, "bj")
	if err != nil {
		return
	}
	cli = c
}

func TestMetric(t *testing.T) {
	startTime := time.Now().AddDate(0, 0, -2)
	endTime := time.Now().AddDate(0, 0, -1)
	request := types.DescribeMetricListRequest{
		MetricName: types.MetricItemCpuIdlePercent,
		Period:     "86400",
		StartTime:  startTime,
		EndTime:    endTime,
		Filter: types.MetricListInstanceFilter{
			InstanceIds: []string{
				"1111",
				"2222",
			},
		},
	}
	metricList, err := cli.DescribeMetricList(context.TODO(), request)
	t.Logf("rsp:%v,err:%v", metricList, err)
}

func TestRegions(t *testing.T) {
	regionsRequest := types.DescribeRegionsRequest{
		ResourceType: types.ResourceTypeInstance,
	}
	regions, err := cli.DescribeRegions(context.TODO(), regionsRequest)
	t.Logf("rsp:%v,err:%v", regions, err)
}

func TestInstances(t *testing.T) {
	instancesRequest := types.DescribeInstancesRequest{}
	instances, err := cli.DescribeInstances(context.TODO(), instancesRequest)
	t.Logf("rsp:%v,err:%v", instances, err)
}

func TestGetAllRegionInstances(t *testing.T) {
	regionsRequest := types.DescribeRegionsRequest{
		ResourceType: types.ResourceTypeInstance,
	}
	regions, err := cli.DescribeRegions(context.TODO(), regionsRequest)
	if err != nil {
		t.Error(err)
	}
	var list []types.ItemDescribeInstance
	for _, i := range regions.List {
		c, err := New(_AK, _SK, i.RegionId)
		if err != nil {
			t.Error(err)
			return
		}
		instancesRequest := types.DescribeInstancesRequest{}
		instances, err := c.DescribeInstances(context.TODO(), instancesRequest)
		if err != nil {
			t.Error(err)
			return
		}
		if len(instances.List) == 0 {
			continue
		}
		list = append(list, instances.List...)
	}
	t.Logf("list:%v", list)
}
