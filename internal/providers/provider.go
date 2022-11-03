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
	ProviderType() string
	QueryAccountBill(ctx context.Context, request types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error)
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
