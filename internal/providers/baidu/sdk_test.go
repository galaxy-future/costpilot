package baidu

import (
	"fmt"
	"testing"
	"time"

	"github.com/galaxy-future/costpilot/internal/providers/types"
)

func TestBceClient_Send(t *testing.T) {
	params := []QueryParam{
		{
			K: "dimensions",
			V: "InstanceId:fakeid-2222-8888-1111-13a8469b1fb2",
		},
		{
			K: "endTime",
			V: time.Now().AddDate(0, -1, 0).Format("2006-01-02T15:04:05Z"),
		},
		{
			K: "periodInSecond",
			V: "60",
		},
		{
			K: "startTime",
			V: time.Now().AddDate(0, -1, -10).Format("2006-01-02T15:04:05Z"),
		},
		{
			K: "statistics[]",
			V: "average,maximum,minimum",
		},
	}
	c := NewBCMClient("", "", "bcm.bj.baidubce.com")
	path := fmt.Sprintf("/json-api/v1/metricdata/%s/%s/%s", "", "BCE_BCC", types.MetricItemCpuIdlePercent)
	rsp, err := c.Send(path, params)
	t.Logf("rsp:%v,err:%v", rsp, err)
}

func TestMetric(t *testing.T) {
	startTime := time.Now().AddDate(0, -1, -3)
	endTime := time.Now().AddDate(0, -1, 0)
	client, _ := New("", "", "bj")
	request := types.DescribeMetricListRequest{MetricName: types.MetricItemCpuIdlePercent, StartTime: startTime, EndTime: endTime}
	metricList, err := client.DescribeMetricList(nil, request)
	t.Logf("rsp:%v,err:%v", metricList, err)
}

func TestRegions(t *testing.T) {
	client, _ := New("", "", "bj")
	regionsRequest := types.DescribeRegionsRequest{ResourceType: types.ResourceTypeInstance}
	regions, err := client.DescribeRegions(nil, regionsRequest)
	t.Logf("rsp:%v,err:%v", regions, err)
}

func TestInstances(t *testing.T) {
	client, _ := New("", "", "bj")
	instancesRequest := types.DescribeInstancesRequest{}
	instances, err := client.DescribeInstances(nil, instancesRequest)
	t.Logf("rsp:%v,err:%v", instances, err)
}
