package baidu

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/baidubce/bce-sdk-go/services/bcc"

	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
)

type BaiduCloud struct {
	bccClient  *bcc.Client
	bcmClient  *BCMClient
	billClient *BCMClient
}

const _billEndPoint = "billing.baidubce.com"

var ProductType = []string{"prepay", "postpay"}

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
		bcmClient:  NewBCMClient(ak, sk, ep),
		billClient: NewBCMClient(ak, sk, _billEndPoint),
	}, nil
}

// ProviderType
func (*BaiduCloud) ProviderType() cloud.Provider {
	return cloud.BaiduCloud
}

// QueryAccountBill
func (p *BaiduCloud) QueryAccountBill(ctx context.Context, param types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error) {
	var (
		params   []QueryParam
		pageNum  = 1
		pageSize = 100
	)
	if param.Granularity == types.Monthly {
		if param.BillingCycle == "" {
			return types.DataInQueryAccountBill{}, errors.New("unknown billing cycle")
		}
		params = append(params, QueryParam{
			K: "month",
			V: param.BillingCycle,
		})
	} else if param.Granularity == types.Daily {
		if param.BillingDate == "" {
			return types.DataInQueryAccountBill{}, errors.New("unknown billing date")
		}
		params = append(params, QueryParam{
			K: "beginTime",
			V: param.BillingDate,
		})
		params = append(params, QueryParam{
			K: "endTime",
			V: param.BillingDate,
		})
	}

	var bill struct {
		BillMonth  string `json:"billMonth,omitempty"`
		AccountId  string `json:"accountId"`
		LoginName  string `json:"loginName"`
		Message    string `json:"message,omitempty"`
		Code       string `json:"code,omitempty"`
		TotalCount int    `json:"totalCount"`
		Bills      []*struct {
			ServiceType       string  `json:"serviceType"`
			ServiceTypeName   string  `json:"serviceTypeName"`
			OrderPurchaseTime string  `json:"orderPurchaseTime"`
			ProductType       string  `json:"productType"`
			FinancePrice      float64 `json:"financePrice"`
		} `json:"bills,omitempty"`
	}

	bills := types.DataInQueryAccountBill{Items: types.ItemsInQueryAccountBill{}}
	tempMap := make(map[string]*types.AccountBillItem)
	billItems := make([]types.AccountBillItem, 0)
	for _, productType := range ProductType {
		pageNum = 1
		for {
			tempParams := params
			tempParams = append(tempParams, QueryParam{
				K: "productType",
				V: productType,
			})
			tempParams = append(tempParams, QueryParam{
				K: "pageNo",
				V: strconv.Itoa(pageNum),
			},
				QueryParam{
					K: "pageSize",
					V: strconv.Itoa(pageSize),
				})
			response, err := p.billClient.Send("/v1/bill/resource/month", tempParams)
			if err != nil {
				return bills, err
			}
			if err = json.Unmarshal(response, &bill); err != nil {
				return bills, err
			}
			if bill.Code != "" {
				return bills, errors.New(bill.Message)
			}
			bills.BillingCycle = bill.BillMonth
			bills.AccountID = bill.AccountId
			bills.TotalCount = bill.TotalCount
			bills.AccountName = bill.LoginName
			if bills.TotalCount == 0 {
				break
			}
			for _, data := range bill.Bills {
				d := types.AccountBillItem{
					PipCode:          convPipCode(data.ServiceType),
					ProductName:      data.ServiceTypeName,
					BillingDate:      data.OrderPurchaseTime,
					SubscriptionType: convSubscriptionType(data.ProductType),
					PretaxAmount:     data.FinancePrice,
				}
				billItems = append(billItems, d)
			}
			if len(billItems) >= bill.TotalCount {
				break
			}
			pageNum++
		}
	}

	if param.IsGroupByProduct {
		for _, data := range billItems {
			if val, ok := tempMap[data.ProductName+data.SubscriptionType.String()]; ok {
				val.PretaxAmount += data.PretaxAmount
			} else {
				tempMap[data.ProductName+data.SubscriptionType.String()] = &data
			}
		}

		billItems = make([]types.AccountBillItem, 0, len(tempMap))
		for _, v := range tempMap {
			billItems = append(billItems, *v)
		}
	}

	bills.Items.Item = billItems
	return bills, nil
}

func (p *BaiduCloud) DescribeMetricList(ctx context.Context, param types.DescribeMetricListRequest) (types.DescribeMetricList, error) {
	// TODO implement me
	panic("implement me")
}

func (p *BaiduCloud) DescribeInstanceAttribute(ctx context.Context, param types.DescribeInstanceAttributeRequest) (types.DescribeInstanceAttribute, error) {
	// TODO implement me
	panic("implement me")
}

func (p *BaiduCloud) DescribeRegions(ctx context.Context, param types.DescribeRegionsRequest) (types.DescribeRegions, error) {
	// TODO implement me
	panic("implement me")
}

func (p *BaiduCloud) DescribeInstanceBill(ctx context.Context, param types.DescribeInstanceBillRequest, isAll bool) (types.DescribeInstanceBill, error) {
	// TODO implement me
	panic("implement me")
}

func (p *BaiduCloud) QueryAvailableInstances(ctx context.Context, param types.QueryAvailableInstancesRequest) (types.QueryAvailableInstances, error) {
	// TODO implement me
	panic("implement me")
}

func convPipCode(pipCode string) types.PipCode {
	switch pipCode {
	// 弹性云服务器
	case "BCC":
		return types.ECS
	// 弹性公网IP
	case "EIP":
		return types.EIP
	// 对象存储服务
	case "BOS":
		return types.S3
	// 云磁盘
	case "CDS":
		return types.CDS
	}
	return types.PipCode(pipCode)
}

func convSubscriptionType(chargeMode string) cloud.SubscriptionType {
	switch chargeMode {
	// 预付费
	case "prepay":
		return cloud.PrePaid
	// 后付费
	case "postpay":
		return cloud.PostPaid
	}
	return "undefined"
}
