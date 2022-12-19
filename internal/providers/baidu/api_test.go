package baidu

import (
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"testing"
)

func TestBceClient_Bill_Send(t *testing.T) {
	request := types.QueryAccountBillRequest{Granularity: types.Daily, BillingDate: "2019-07-19"}
	cloud, _ := New("", "", "bj")
	rsp, err := cloud.QueryAccountBill(nil, request)
	t.Logf("rsp:%v,err:%v", rsp, err)
}
