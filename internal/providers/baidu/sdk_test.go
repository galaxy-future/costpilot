package baidu

import (
	"fmt"
	"testing"
)

func TestBceClient_Send(t *testing.T) {
	params := []QueryParam{
		{
			K: "dimensions",
			V: "InstanceId:fakeid-2222-8888-1111-13a8469b1fb2",
		},
		{
			K: "endTime",
			V: "2022-11-24T00:00:00Z",
		},
		{
			K: "periodInSecond",
			V: "3600",
		},
		{
			K: "startTime",
			V: "2022-11-23T00:00:00Z",
		},
		{
			K: "statistics[]",
			V: "average",
		},
	}
	c := NewBCMClient("xx", "xx", "bcm.bj.baidubce.com")
	path := fmt.Sprintf("/json-api/v1/metricdata/%s/%s/%s", "41aecd6690764a28a3c737fc554f017c", "BCE_BCC", "MemUsedPercent")
	rsp, err := c.Send(path, params)
	t.Logf("rsp:%v,err:%v", rsp, err)
}
