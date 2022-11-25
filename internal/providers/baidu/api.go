package baidu

import (
	"context"
	"strings"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/pkg/errors"
)

type BaiduCloud struct {
	bccClient *bcc.Client
	bcmClient *BCMClient
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
		bccClient: bccClient,
		bcmClient: NewBCMClient(ak, sk, ep),
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

// DescribeMetricList
// 在使用 BCMClient.Send 方法请求时，注意参数顺序，参看 TestBceClient_Send
func (p *BaiduCloud) DescribeMetricList(ctx context.Context, param types.DescribeMetricListRequest) (types.DescribeMetricList, error) {
	// TODO implement me
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
