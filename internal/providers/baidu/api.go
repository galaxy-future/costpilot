package baidu

import (
	"context"
	"net/http"
	"strings"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/pkg/errors"
)

type BaiduCloud struct {
	bccClient  *bcc.Client
	httpClient *http.Client
}

var EndPoints = map[string]string{
	"bj":  ".bj.baidubce.com",
	"gz":  ".gz.baidubce.com",
	"su":  ".su.baidubce.com",
	"hkg": ".hkg.baidubce.com",
	"fwh": ".fwh.baidubce.com",
	"bd":  ".bd.baidubce.com",
}

func New(ak, sk, regionId string) (*BaiduCloud, error) {
	ep, ok := EndPoints[strings.ToLower(regionId)]
	if !ok {
		return nil, errors.New("regionId error:" + regionId)
	}

	bccClient, err := bcc.NewClient(ak, sk, ep)
	if err != nil {
		return nil, err
	}

	return &BaiduCloud{
		bccClient:  bccClient,
		httpClient: &http.Client{},
	}, nil
}

// ProviderType
func (*BaiduCloud) ProviderType() cloud.Provider {
	return cloud.BaiduCloud
}

// QueryAccountBill
func (p *BaiduCloud) QueryAccountBill(ctx context.Context, param types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error) {
	return types.DataInQueryAccountBill{}, nil
}

func (p *BaiduCloud) DescribeMetricList(ctx context.Context, param types.DescribeMetricListRequest) (types.DescribeMetricList, error) {
	// TODO implement me
	// 百度云没有提供官方 sdk，但我们可以自己组装参数，用 go http 发起请求，实现参见：https://github.com/galaxy-future/bridgx/blob/dev/pkg/cloud/baidu/cr.go#L37
	return types.DescribeMetricList{}, nil
}

func (p *BaiduCloud) DescribeRegions(ctx context.Context, param types.DescribeRegionsRequest) (types.DescribeRegions, error) {
	// TODO implement me
	return types.DescribeRegions{}, nil
}

func (p *BaiduCloud) DescribeInstanceBill(ctx context.Context, param types.DescribeInstanceBillRequest, isAll bool) (types.DescribeInstanceBill, error) {
	return types.DescribeInstanceBill{}, nil
}

func (p *BaiduCloud) QueryAvailableInstances(ctx context.Context, param types.QueryAvailableInstancesRequest) (types.QueryAvailableInstances, error) {
	return types.QueryAvailableInstances{}, nil
}

func (p *BaiduCloud) DescribeInstances(ctx context.Context, param types.DescribeInstancesRequest) (types.DescribeInstances, error) {
	// TODO implement me
	return types.DescribeInstances{}, nil
}
