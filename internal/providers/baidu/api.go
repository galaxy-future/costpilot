package baidu

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/galaxy-future/costpilot/internal/services"
	"github.com/pkg/errors"
)

type BaiduCloud struct {
	ak        string
	bccClient *bcc.Client
	bcmClient *BCMClient
}

var (
	_metricNameMap = map[types.MetricItem]string{
		types.MetricItemMemUsedPercent: "MemUsedPercent",
		types.MetricItemCpuIdlePercent: "CpuIdlePercent",
	}
)

func New(ak, sk, regionId string) (*BaiduCloud, error) {
	ep, ok := _endPointMap[strings.ToLower(regionId)]
	if !ok {
		return nil, errors.New("regionId error:" + regionId)
	}

	bccClient, err := bcc.NewClient(ak, sk, fmt.Sprintf("bcc%s", ep))
	if err != nil {
		return nil, err
	}
	return &BaiduCloud{
		ak:        ak,
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
// https://cloud.baidu.com/doc/BCM/s/9jwvym3kb
func (p *BaiduCloud) DescribeMetricList(ctx context.Context, param types.DescribeMetricListRequest) (types.DescribeMetricList, error) {
	metricName, ok := _metricNameMap[param.MetricName]
	if !ok {
		return types.DescribeMetricList{}, errors.New("unknown metric name")
	}
	if len(param.Filter.InstanceIds) == 0 {
		return types.DescribeMetricList{}, errors.New("unknown instance id")
	}
	accountId, err := p.getAccountId()
	if err != nil {
		return types.DescribeMetricList{}, err
	}
	var metricList []types.MetricSample
	for _, id := range param.Filter.InstanceIds {
		// 当监控项具备多个维度时使用分号连接，例如dimensionName:dimensionValue;dimensionName:dimensionValue，相同维度只能指定一个维度值
		queryParam := []QueryParam{
			{K: "dimensions", V: fmt.Sprintf("InstanceId:%s", id)},
			{K: "endTime", V: param.EndTime.Format("2006-01-02T15:04:05Z")},
			{K: "periodInSecond", V: param.Period},
			{K: "startTime", V: param.StartTime.Format("2006-01-02T15:04:05Z")},
			{K: "statistics[]", V: "average,maximum,minimum"},
		}
		metric, err := p.getMetricByInstanceId(ctx, queryParam, id, metricName, accountId)
		if err != nil {
			return types.DescribeMetricList{}, err
		}
		metricList = append(metricList, metric...)
	}
	return types.DescribeMetricList{List: metricList}, nil
}

func (p *BaiduCloud) getMetricByInstanceId(_ context.Context, queryParam []QueryParam, instanceId, metricName, accountId string) ([]types.MetricSample, error) {
	var ret []types.MetricSample
	path := fmt.Sprintf("/json-api/v1/metricdata/%s/%s/%s", accountId, "BCE_BCC", metricName)
	response, err := p.bcmClient.Send(path, queryParam)
	if err != nil {
		return ret, err
	}
	if response["code"] != "OK" {
		return ret, fmt.Errorf("%s", response["message"])
	}

	var dataList []*struct {
		Timestamp time.Time `json:"timestamp"`
		Minimum   float64   `json:"minimum"`
		Maximum   float64   `json:"maximum"`
		Average   float64   `json:"average"`
	}
	bytes, err := json.Marshal(response["dataPoints"])
	if err != nil {
		return ret, err
	}
	if err = json.Unmarshal(bytes, &dataList); err != nil {
		return ret, nil
	}

	for _, datapoint := range dataList {
		d := types.MetricSample{
			InstanceId: instanceId,
			Min:        datapoint.Minimum,
			Max:        datapoint.Maximum,
			Average:    datapoint.Average,
			Timestamp:  datapoint.Timestamp.Unix(),
		}
		ret = append(ret, d)
	}
	return ret, nil
}

func (p *BaiduCloud) getAccountId() (string, error) {
	accounts := services.NewAccountService().GetAccounts()
	if len(accounts) == 0 {
		return "", errors.New("BaiduCloud account id is not configured")
	}
	for _, a := range accounts {
		if a.Provider == cloud.BaiduCloud && a.AK == p.ak {
			return a.AccountID, nil
		}
	}

	return "", errors.New("BaiduCloud account id is not configured")
}

func (p *BaiduCloud) DescribeRegions(_ context.Context, param types.DescribeRegionsRequest) (types.DescribeRegions, error) {
	if param.ResourceType == "" {
		return types.DescribeRegions{}, errors.New("unknown resource type")
	}
	var regionList []types.ItemRegion
	for regionId, name := range _regionNameMap {
		regionList = append(regionList, types.ItemRegion{
			LocalName: name,
			RegionId:  regionId,
		})
	}
	return types.DescribeRegions{List: regionList}, nil
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
		items = append(items, convInstance(response.Instances)...)
		if !response.IsTruncated {
			break
		}
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
			SubscriptionType: convSubscriptionType(item.PaymentTiming),
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
