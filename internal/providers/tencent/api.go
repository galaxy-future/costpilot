package tencent

import (
	"context"

	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
)

type TencentCloud struct {
}

func New(ak, sk, regionId string) (*TencentCloud, error) {

	return &TencentCloud{}, nil
}

// ProviderType
func (*TencentCloud) ProviderType() string {
	return cloud.TencentCloud
}

// QueryAccountBill
func (p *TencentCloud) QueryAccountBill(ctx context.Context, param types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error) {

	return types.DataInQueryAccountBill{}, nil
}
