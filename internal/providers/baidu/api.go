package baidu

import (
	"context"

	"github.com/galayx-future/costpilot/internal/constants/cloud"
	"github.com/galayx-future/costpilot/internal/providers/types"
)

type BaiduCloud struct {
}

func New(ak, sk, regionId string) (*BaiduCloud, error) {

	return &BaiduCloud{}, nil
}

// ProviderType
func (*BaiduCloud) ProviderType() string {
	return cloud.BaiduCloud
}

// QueryAccountBill
func (p *BaiduCloud) QueryAccountBill(ctx context.Context, param types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error) {

	return types.DataInQueryAccountBill{}, nil
}
