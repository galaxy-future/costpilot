package baidu

import (
	"context"

	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
)

type BaiduCloud struct {
}

func New(ak, sk, regionId string) (*BaiduCloud, error) {

	return &BaiduCloud{}, nil
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
	panic("implement me")
}

func (p *BaiduCloud) DescribeInstanceAttribute(ctx context.Context, request types.DescribeInstanceAttributeRequest) (types.DescribeInstanceAttribute, error) {
	// TODO implement me
	panic("implement me")
}

func (p *BaiduCloud) DescribeRegions(ctx context.Context, request types.DescribeRegionsRequest) (types.DescribeRegions, error) {
	// TODO implement me
	panic("implement me")
}
