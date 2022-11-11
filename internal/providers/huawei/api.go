package huawei

import (
	"context"

	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
)

type HuaweiCloud struct {
}

func New(ak, sk, regionId string) (*HuaweiCloud, error) {

	return &HuaweiCloud{}, nil
}

// ProviderType
func (*HuaweiCloud) ProviderType() cloud.Provider {
	return cloud.HuaweiCloud
}

// QueryAccountBill
func (p *HuaweiCloud) QueryAccountBill(ctx context.Context, param types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error) {

	return types.DataInQueryAccountBill{}, nil
}

func (p *HuaweiCloud) DescribeMetricList(ctx context.Context, param types.DescribeMetricListRequest) (types.DescribeMetricList, error) {
	return types.DescribeMetricList{}, nil
}

func (p *HuaweiCloud) DescribeInstanceAttribute(ctx context.Context, param types.DescribeInstanceAttributeRequest) (types.DescribeInstanceAttribute, error) {
	// TODO implement me
	panic("implement me")
}

func (p *HuaweiCloud) DescribeRegions(ctx context.Context, param types.DescribeRegionsRequest) (types.DescribeRegions, error) {
	panic("implement me")
}

func (p *HuaweiCloud) DescribeInstanceBill(ctx context.Context, param types.DescribeInstanceBillRequest, isAll bool) (types.DescribeInstanceBill, error) {
	// TODO implement me
	panic("implement me")
}

func (p *HuaweiCloud) QueryAvailableInstances(ctx context.Context, param types.QueryAvailableInstancesRequest) (types.QueryAvailableInstances, error) {
	// TODO implement me
	panic("implement me")
}
