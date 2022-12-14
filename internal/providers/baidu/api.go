package baidu

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/galaxy-future/costpilot/internal/services"

	"github.com/baidubce/bce-sdk-go/services/bcc/api"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/pkg/errors"
)

type BaiduCloud struct {
	bccClient *bcc.Client
	bcmClient *BCMClient
}

var (
	endPoints = map[string]string{
		"bj":  ".bj.baidubce.com",
		"gz":  ".gz.baidubce.com",
		"su":  ".su.baidubce.com",
		"hkg": ".hkg.baidubce.com",
		"fwh": ".fwh.baidubce.com",
		"bd":  ".bd.baidubce.com",
	}
	metricNameMap = map[types.MetricItem]string{
		types.MetricItemMemUsedPercent: "MemUsedPercent",
		types.MetricItemCpuIdlePercent: "CpuIdlePercent",
	}
)

func New(ak, sk, regionId string) (*BaiduCloud, error) {
	ep, ok := endPoints[strings.ToLower(regionId)]
	if !ok {
		return nil, errors.New("regionId error:" + regionId)
	}

	bccClient, err := bcc.NewClient(ak, sk, fmt.Sprintf("bcc%s", ep))
	if err != nil {
		return nil, err
	}
	return &BaiduCloud{
		bccClient: bccClient,
		bcmClient: NewBCMClient(ak, sk, fmt.Sprintf("bcm%s", ep)),
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
func (p *BaiduCloud) DescribeMetricList(_ context.Context, param types.DescribeMetricListRequest) (types.DescribeMetricList, error) {
	metricName, ok := metricNameMap[param.MetricName]
	if !ok {
		return types.DescribeMetricList{}, errors.New("unknown metric name")
	}
	if len(param.Filter.InstanceIds) == 0 {
		return types.DescribeMetricList{}, errors.New("unknown instance id")
	}
	accounts := services.NewAccountService().GetAccounts()
	if len(accounts) == 0 {
		return types.DescribeMetricList{}, errors.New("cloud account is not configured")
	}
	if accounts[0].Provider != cloud.BaiduCloud {
		return types.DescribeMetricList{}, errors.New("unknown provider")
	}
	params := []QueryParam{
		{
			K: "dimensions",
			V: "InstanceId:" + strings.Join(param.Filter.InstanceIds, ","),
		},
		{
			K: "endTime",
			V: param.EndTime.Format("2006-01-02T15:04:05Z"),
		},
		{
			K: "periodInSecond",
			V: param.Period,
		},
		{
			K: "startTime",
			V: param.StartTime.Format("2006-01-02T15:04:05Z"),
		},
		{
			K: "statistics[]",
			V: "average,maximum,minimum",
		},
	}
	path := fmt.Sprintf("/json-api/v1/metricdata/%s/%s/%s", accounts[0].AccountID, "BCE_BCC", metricName)
	response, err := p.bcmClient.Send(path, params)
	if err != nil {
		return types.DescribeMetricList{}, err
	}
	if response["code"] != "OK" {
		return types.DescribeMetricList{}, fmt.Errorf("%s", response["message"])
	}

	var dataList []*struct {
		Timestamp time.Time `json:"timestamp"`
		Minimum   float64   `json:"minimum"`
		Maximum   float64   `json:"maximum"`
		Average   float64   `json:"average"`
	}
	bytes, err := json.Marshal(response["dataPoints"])
	if err != nil {
		return types.DescribeMetricList{}, nil
	}
	if err = json.Unmarshal(bytes, &dataList); err != nil {
		return types.DescribeMetricList{}, nil
	}

	metricList := types.DescribeMetricList{List: make([]types.MetricSample, 0, len(dataList))}
	for _, datapoint := range dataList {
		d := types.MetricSample{
			Min:       datapoint.Minimum,
			Max:       datapoint.Maximum,
			Average:   datapoint.Average,
			Timestamp: datapoint.Timestamp.Unix(),
		}
		metricList.List = append(metricList.List, d)
	}
	return metricList, nil
}

func (p *BaiduCloud) DescribeRegions(_ context.Context, param types.DescribeRegionsRequest) (types.DescribeRegions, error) {
	if param.ResourceType == "" {
		return types.DescribeRegions{}, errors.New("unknown resource type")
	}
	args := &api.ListTypeZonesArgs{InstanceType: string(param.ResourceType)}
	response, err := p.bccClient.ListTypeZones(args)
	if err != nil {
		return types.DescribeRegions{}, err
	}
	region := types.DescribeRegions{}
	zoneNames := response.ZoneNames
	if len(zoneNames) == 0 {
		return region, nil
	}
	for _, z := range zoneNames {
		region.List = append(region.List, types.ItemRegion{
			LocalName: z,
		})
	}
	return region, nil
}

func (p *BaiduCloud) DescribeInstanceBill(ctx context.Context, param types.DescribeInstanceBillRequest, isAll bool) (types.DescribeInstanceBill, error) {
	return types.DescribeInstanceBill{}, nil
}

func (p *BaiduCloud) QueryAvailableInstances(ctx context.Context, param types.QueryAvailableInstancesRequest) (types.QueryAvailableInstances, error) {
	return types.QueryAvailableInstances{}, nil
}

func (p *BaiduCloud) DescribeInstances(_ context.Context, param types.DescribeInstancesRequest) (types.DescribeInstances, error) {
	listArgs := &api.ListServerRequestV3Args{}
	if len(param.InstanceIds) > 0 {
		listArgs.InstanceId = strings.Join(param.InstanceIds, ",")
	}
	instances := types.DescribeInstances{}
	var items []types.ItemDescribeInstance
	for {
		response, err := p.bccClient.ListServersByMarkerV3(listArgs)
		if err != nil {
			return instances, err
		}
		if !response.IsTruncated {
			break
		}
		items = append(items, convInstance(response.Instances)...)
		listArgs.Marker = response.NextMarker
	}

	instances.List = items
	return instances, nil
}

func convInstance(instances []api.InstanceModelV3) []types.ItemDescribeInstance {
	result := make([]types.ItemDescribeInstance, 0, len(instances))
	for _, item := range instances {
		result = append(result, types.ItemDescribeInstance{
			InstanceId:       item.InstanceId,
			InstanceName:     item.InstanceName,
			RegionId:         item.ZoneName,
			PublicIpAddress:  item.PublicIpAddress,
			InnerIpAddress:   item.PrivateIpAddress,
			SubscriptionType: convSubscriptionType(item.PaymentTiming),
			Status:           string(item.Status),
		})
	}
	return result
}

func convSubscriptionType(subscriptionType string) cloud.SubscriptionType {
	switch subscriptionType {
	case "Prepaid":
		return cloud.PrePaid
	case "Postpaid":
		return cloud.PostPaid
	default:
		return cloud.Undefined
	}
}
