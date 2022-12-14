package baidu

import (
	"fmt"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"testing"
	"time"
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
	c := NewBCMClient("348b76788202435a977bc7a2facaa3ca", "5824df6026bd4dada215b0528f67b04f", "bcm.bj.baidubce.com")
	path := fmt.Sprintf("/json-api/v1/metricdata/%s/%s/%s", "749e1e962f2f4e629ecc1ff3f8801f6b", "BCE_BCC", "CpuIdlePercent")
	rsp, err := c.Send(path, params)
	t.Logf("rsp:%v,err:%v", rsp, err)
	// path = fmt.Sprintf("/json-api/v1/metricdata/%s/%s/%s", "749e1e962f2f4e629ecc1ff3f8801f6b", "BCE_BCC", "CpuIdlePercent")
	// rsp, err = c.Send(path, params)
	// t.Logf("rsp:%v,err:%v", rsp, err)
}

func TestBCC(t *testing.T) {
	// params := []QueryParam{
	// {
	// 	K: "maxKeys",
	// 	V: "1000",
	// },
	// }
	timepoint := time.Now().AddDate(0, -1, 0)
	starttime := time.Now().AddDate(0, -1, -3)
	fmt.Println(timepoint)
	client, _ := New("348b76788202435a977bc7a2facaa3ca", "5824df6026bd4dada215b0528f67b04f", "bj")
	request := types.DescribeMetricListRequest{MetricName: types.MetricItemCpuIdlePercent, StartTime: starttime, EndTime: timepoint}
	_, err := client.DescribeMetricList(nil, request)
	fmt.Println(err)
	regionsRequest := types.DescribeRegionsRequest{ResourceType: types.ResourceTypeInstance}
	client.DescribeRegions(nil, regionsRequest)
	instancesRequest := types.DescribeInstancesRequest{}
	client.DescribeInstances(nil, instancesRequest)
	// listArgs := &api.ListServerRequestV3Args{}
	//
	// v3Instances, err := client.bccClient.ListServersByMarkerV3(listArgs)
	// t.Logf("rsp:%v,err:%v", v3Instances, err)
	//
	// args := &api.ListTypeZonesArgs{
	// 	InstanceType: "N1",
	// 	ProductType:  "",
	// 	Spec:         "",
	// 	SpecId:       "",
	// }
	//
	// listTypeZones, err := client.bccClient.ListTypeZones(args)
	// t.Logf("rsp:%v,err:%v", listTypeZones, err)

	// c := NewBCMClient("348b76788202435a977bc7a2facaa3ca", "5824df6026bd4dada215b0528f67b04f", "bcc.bj.baidubce.com")
	// path := fmt.Sprintf("/v2/instance")
	//
	// rsp, err := c.Send(path, params)
	// t.Logf("rsp:%v,err:%v", rsp, err)

}
