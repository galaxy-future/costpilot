package providers

import (
	"context"
	"fmt"
	"sync"

	"github.com/galaxy-future/costpilot/internal/providers/aws"
	"github.com/galaxy-future/costpilot/internal/providers/baidu"
	"github.com/galaxy-future/costpilot/internal/providers/huawei"
	"github.com/galaxy-future/costpilot/internal/providers/tencent"

	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/alibaba"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/spf13/cast"
)

var clientMap sync.Map

type Provider interface {
	ProviderType() cloud.Provider

	QueryAccountBill(ctx context.Context, request types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error)
	// DescribeInstanceBill query the consumption of all product instances or billing items for a certain account period
	// in principle, we can get the basic info about specify InstantId, even through deleted or released.
	DescribeInstanceBill(ctx context.Context, request types.DescribeInstanceBillRequest, isAll bool) (types.DescribeInstanceBill, error)
	// QueryAvailableInstances list all available Instances by RegionId OR InstantIds.
	QueryAvailableInstances(context.Context, types.QueryAvailableInstancesRequest) (types.QueryAvailableInstances, error)

	// DescribeRegions list all regions as the RegionId and RegionName map.
	DescribeRegions(context.Context, types.DescribeRegionsRequest) (types.DescribeRegions, error)
	// DescribeInstanceAttribute get the Instance detail by only InstantId.
	DescribeInstanceAttribute(context.Context, types.DescribeInstanceAttributeRequest) (types.DescribeInstanceAttribute, error)

	// DescribeMetricList get monitoring samples, eg: cpu/memory.
	DescribeMetricList(context.Context, types.DescribeMetricListRequest) (types.DescribeMetricList, error)
}

// GetProvider get provider
func GetProvider(provider cloud.Provider, ak, sk, regionID string) (Provider, error) {
	var client Provider
	var err error
	key := cast.ToString(provider) + ak + regionID
	v, exist := clientMap.Load(key)
	if exist {
		return v.(Provider), nil
	}

	switch provider {
	case cloud.AlibabaCloud:
		client, err = alibaba.New(ak, sk, regionID)
	case cloud.HuaweiCloud:
		client, err = huawei.New(ak, sk, regionID)
	case cloud.AWSCloud:
		client, err = aws.New(ak, sk, regionID)
	case cloud.TencentCloud:
		client, err = tencent.New(ak, sk, regionID)
	case cloud.BaiduCloud:
		client, err = baidu.New(ak, sk, regionID)
	default:
		return nil, fmt.Errorf("invalid provider[%s]", provider)
	}
	if err != nil {
		return nil, err
	}
	clientMap.Store(key, client)
	return client, nil
}
