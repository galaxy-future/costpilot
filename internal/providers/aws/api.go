package aws

import (
	"context"
	"errors"
	"fmt"
	cloudwatchType "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"log"
	"strconv"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/galaxy-future/costpilot/tools"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	explorerTypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/providers/types"
)

type AWSCloud struct {
	client     *costexplorer.Client
	ec2Client  *ec2.Client
	cloudWatch *cloudwatch.Client
}

func New(AK, SK, regionId string) (*AWSCloud, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(regionId), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(AK, SK, "")))
	if err != nil {
		return nil, err
	}

	return &AWSCloud{
		client:     costexplorer.NewFromConfig(cfg),
		ec2Client:  ec2.NewFromConfig(cfg),
		cloudWatch: cloudwatch.NewFromConfig(cfg),
	}, nil
}

// ProviderType
func (*AWSCloud) ProviderType() cloud.Provider {
	return cloud.AWSCloud
}

// QueryAccountBill
func (p *AWSCloud) QueryAccountBill(ctx context.Context, param types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error) {
	items := make([]types.AccountBillItem, 0)
	var err error
	if param.IsGroupByProduct {
		result1, err := p.QueryByFilter(param, "On Demand Instances")
		if err != nil {
			return types.DataInQueryAccountBill{}, err
		}
		result2, err := p.QueryByFilter(param, "Standard Reserved Instances")
		if err != nil {
			return types.DataInQueryAccountBill{}, err
		}
		items = append(result1, result2...)
	} else {
		items, err = p.QueryByFilter(param, "")
		if err != nil {
			return types.DataInQueryAccountBill{}, err
		}
	}
	result := types.DataInQueryAccountBill{
		BillingCycle: param.BillingCycle,
		TotalCount:   len(items),
		Items: types.ItemsInQueryAccountBill{
			Item: items,
		},
	}
	return result, nil
}

func convGranularity(granularity types.Granularity) explorerTypes.Granularity {
	switch granularity {
	case types.Daily:
		return explorerTypes.GranularityDaily
	case types.Monthly:
		return explorerTypes.GranularityMonthly
	default:
		return ""
	}
}

// convert the costexplorer.GetCostAndUsageOutput to types.AccountBillItem
func convAccountBillItems(output *costexplorer.GetCostAndUsageOutput, param types.QueryAccountBillRequest, chargeType string) ([]types.AccountBillItem, error) {
	var result []types.AccountBillItem
	var err error
	resultBytime := output.ResultsByTime[0]
	if param.IsGroupByProduct {
		result = make([]types.AccountBillItem, 0, len(resultBytime.Groups))
		for _, group := range resultBytime.Groups {
			newItem := types.AccountBillItem{
				SubscriptionType: convChargeType(chargeType),
				PipCode:          types.PipCode(group.Keys[0]),
				Currency:         aws.StringValue(group.Metrics["BlendedCost"].Unit),
				ProductName:      group.Keys[0],
			}
			newItem.PretaxAmount, err = convAmount(group.Metrics["BlendedCost"].Amount)
			if err != nil {
				return nil, err
			}
			if param.Granularity == types.Daily {
				newItem.BillingDate = param.BillingDate
			}
			result = append(result, newItem)
		}
	} else {
		result = make([]types.AccountBillItem, 0, 1)
		newItem := types.AccountBillItem{
			SubscriptionType: convChargeType(chargeType),
			Currency:         aws.StringValue(resultBytime.Total["BlendedCost"].Unit),
		}
		newItem.PretaxAmount, err = convAmount(resultBytime.Total["BlendedCost"].Amount)
		if err != nil {
			return nil, err
		}
		if param.Granularity == types.Daily {
			newItem.BillingDate = param.BillingDate
		}
		result = append(result, newItem)
	}
	return result, nil
}

func convAmount(amountPtr *string) (float64, error) {
	if amountPtr == nil {
		return 0, errors.New("amountPtr is nil")
	}
	t := tea.StringValue(amountPtr)
	return strconv.ParseFloat(t, 64)
}

func convChargeType(s string) (result cloud.SubscriptionType) {
	switch s {
	case "On Demand Instances":
		result = cloud.PostPaid
	case "Standard Reserved Instances":
		result = cloud.PrePaid
	}
	return
}

func (p *AWSCloud) QueryByFilter(param types.QueryAccountBillRequest, chargeType string) ([]types.AccountBillItem, error) {
	const dateFormat string = "2006-01-02"
	var start, end time.Time
	var err error
	var billItems []types.AccountBillItem
	if param.Granularity == types.Monthly {
		if !IsValidMonth(param.BillingCycle) {
			return []types.AccountBillItem{}, nil
		}
		start, err = time.Parse(dateFormat, param.BillingCycle+"-01")
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
		end = tools.AddDate(start, 0, 1, 0)
		if err != nil {
			return nil, err
		}
	} else if param.Granularity == types.Daily {
		if !IsValidDate(param.BillingDate) {
			return []types.AccountBillItem{}, nil
		}
		start, err = time.Parse(dateFormat, param.BillingDate)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
		end = start.AddDate(0, 0, 1)
		if err != nil {
			return nil, err
		}
	}
	input := &costexplorer.GetCostAndUsageInput{
		Granularity: convGranularity(param.Granularity),
		Metrics:     []string{"BlendedCost"},
		TimePeriod: &explorerTypes.DateInterval{
			End:   aws.String(end.Format(dateFormat)),
			Start: aws.String(start.Format(dateFormat)),
		},
	}

	if param.IsGroupByProduct {
		input.GroupBy = []explorerTypes.GroupDefinition{
			{
				Key:  aws.String(string(explorerTypes.DimensionService)),
				Type: explorerTypes.GroupDefinitionTypeDimension,
			},
		}
	}
	if chargeType != "" {
		input.Filter = &explorerTypes.Expression{
			Dimensions: &explorerTypes.DimensionValues{
				Key:    explorerTypes.DimensionPurchaseType,
				Values: []string{chargeType},
			},
		}
	}
	for {
		output, err := p.client.GetCostAndUsage(context.Background(), input)
		if err != nil {
			return nil, err
		}
		newbillItems, err := convAccountBillItems(output, param, chargeType)
		if err != nil {
			return nil, err
		}
		billItems = append(billItems, newbillItems...)
		if output != nil && output.NextPageToken != nil {
			input.NextPageToken = output.NextPageToken
		} else {
			break
		}
	}
	return billItems, nil

}
func IsValidDate(date string) bool {
	t, _ := time.Parse("2006-01-02", date)
	if t.Before(tools.AddDate(time.Now(), -1, 0, 0)) {
		return false
	}
	return true
}
func IsValidMonth(month string) bool {
	t, _ := time.Parse("2006-01", month)
	if t.Before(tools.AddDate(time.Now(), -1, 0, 0)) {
		return false
	}
	return true
}
func (p *AWSCloud) DescribeRegions(ctx context.Context, param types.DescribeRegionsRequest) (types.DescribeRegions, error) {
	input := &ec2.DescribeRegionsInput{}
	response, err := p.ec2Client.DescribeRegions(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Println(aerr.Error())
			}
		} else {
			log.Println(err.Error())
		}
		return types.DescribeRegions{}, err
	}
	if response.Regions != nil {
		itemRegions := make([]types.ItemRegion, 0, len(response.Regions))
		for _, regin := range response.Regions {
			newRegion := types.ItemRegion{
				RegionId:  aws.StringValue(regin.RegionName),
				LocalName: _regionLocalName[aws.StringValue(regin.RegionName)],
			}
			itemRegions = append(itemRegions, newRegion)
		}
		return types.DescribeRegions{
			List: itemRegions,
		}, err
	}
	return types.DescribeRegions{}, err
}

func (p *AWSCloud) DescribeInstances(ctx context.Context, param types.DescribeInstancesRequest) (types.DescribeInstances, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: param.InstanceIds,
	}
	output, err := p.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Println(aerr.Error())
			}
		} else {
			log.Println(err.Error())
		}
	}
	if output.Reservations != nil {
		reservedInstances, err1 := p.describeReservedInstances(ctx)
		if err1 != nil {
			log.Println(err1.Error())
			return types.DescribeInstances{}, err1
		}
		return convDescribeInstances(output.Reservations, reservedInstances), err
	}
	return types.DescribeInstances{}, err
}

//Get Reserved Instances
func (p *AWSCloud) describeReservedInstances(ctx context.Context) (map[string]string, error) {
	reservedInstances := make(map[string]string, 0)
	output, err := p.ec2Client.DescribeReservedInstances(ctx, &ec2.DescribeReservedInstancesInput{})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Println(aerr.Error())
			}
		} else {
			log.Println(err.Error())
		}
		return reservedInstances, err
	}
	if output.ReservedInstances != nil {
		for _, reservation := range output.ReservedInstances {
			if ec2Types.ReservedInstanceStateActive == reservation.State {
				reservedInstances[aws.StringValue(reservation.ReservedInstancesId)] = string(reservation.InstanceType)
			}
		}
	}
	return reservedInstances, err
}

func convDescribeInstances(reservations []ec2Types.Reservation, reservedInstances map[string]string) types.DescribeInstances {
	awsInstances := make([]types.ItemDescribeInstance, 0)
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			region := aws.StringValue(instance.Placement.AvailabilityZone)
			subscriptionType := cloud.PostPaid
			for k, v := range reservedInstances {
				if string(instance.InstanceType) == v {
					subscriptionType = cloud.PrePaid
					delete(reservedInstances, k)
				}
			}
			newInstance := types.ItemDescribeInstance{
				InstanceId:       aws.StringValue(instance.InstanceId),
				InstanceName:     aws.StringValue(instance.Tags[0].Value),
				RegionId:         aws.StringValue(instance.Placement.AvailabilityZone),
				RegionName:       _regionLocalName[region[0:len(region)-1]],
				SubscriptionType: subscriptionType,
				PublicIpAddress:  []string{aws.StringValue(instance.PublicIpAddress)},
				InnerIpAddress:   []string{aws.StringValue(instance.PrivateIpAddress)},
			}
			awsInstances = append(awsInstances, newInstance)
		}
	}
	return types.DescribeInstances{
		TotalCount: len(awsInstances),
		List:       awsInstances,
	}
}

func convDescribeMetricListRequest(param types.DescribeMetricListRequest) (*cloudwatch.GetMetricDataInput, map[string]string) {
	ids := make(map[string]string)
	metricDataQueries := []cloudwatchType.MetricDataQuery{}
	var nameSpace, metricName, label string
	if types.MetricItemCPUUtilization == param.MetricName {
		nameSpace = Namespace_Cpu
		metricName = CPUUtilization
		label = CPUUtilization
	}
	if types.MetricItemMemoryUsedUtilization == param.MetricName {
		nameSpace = Namespace_Mem
		metricName = MemoryUtilization
		label = MemoryUtilization
	}
	period, _ := strconv.Atoi(param.Period)
	for i, instanceId := range param.Filter.InstanceIds {
		dimension := cloudwatchType.Dimension{
			Name:  aws.String(InstanceId),
			Value: aws.String(instanceId),
		}
		metricDataQuery := cloudwatchType.MetricDataQuery{
			Id: aws.String(fmt.Sprintf("%s%s", "instance", strconv.Itoa(i))),
			MetricStat: &cloudwatchType.MetricStat{
				Metric: &cloudwatchType.Metric{
					Namespace:  aws.String(nameSpace),
					MetricName: aws.String(metricName),
					Dimensions: []cloudwatchType.Dimension{dimension},
				},
				Stat:   aws.String(string(cloudwatchType.StatisticAverage)),
				Period: aws.Int32(int32(period)),
			},
			Label: aws.String(label),
		}
		metricDataQueries = append(metricDataQueries, metricDataQuery)
		ids[aws.StringValue(metricDataQuery.Id)] = instanceId
	}
	input := &cloudwatch.GetMetricDataInput{
		StartTime:         aws.Time(param.StartTime),
		EndTime:           aws.Time(param.EndTime),
		MetricDataQueries: metricDataQueries,
	}
	return input, ids
}
func (p *AWSCloud) DescribeMetricList(ctx context.Context, param types.DescribeMetricListRequest) (types.DescribeMetricList, error) {
	if param.Filter.InstanceIds == nil || len(param.Filter.InstanceIds) == 0 {
		return types.DescribeMetricList{}, nil
	}
	request, ids := convDescribeMetricListRequest(param)
	output, err := p.cloudWatch.GetMetricData(ctx, request)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Println(aerr.Error())
			}
		} else {
			log.Println(err.Error())
		}
		return types.DescribeMetricList{}, err
	}
	if output.MetricDataResults != nil {
		list := make([]types.MetricSample, 0)
		for _, metricDataResult := range output.MetricDataResults {
			if len(metricDataResult.Values) > 0 {
				for i, value := range metricDataResult.Values {
					metricSample := types.MetricSample{
						InstanceId: ids[aws.StringValue(metricDataResult.Id)],
						Average:    value,
						Timestamp:  aws.TimeUnixMilli(metricDataResult.Timestamps[i]),
					}
					list = append(list, metricSample)
				}
			}

		}
		return types.DescribeMetricList{
			List: list,
		}, err
	}
	return types.DescribeMetricList{}, err
}

func (p *AWSCloud) DescribeInstanceBill(ctx context.Context, param types.DescribeInstanceBillRequest, isAll bool) (types.DescribeInstanceBill, error) {
	return types.DescribeInstanceBill{}, nil
}

func (p *AWSCloud) QueryAvailableInstances(ctx context.Context, param types.QueryAvailableInstancesRequest) (types.QueryAvailableInstances, error) {
	return types.QueryAvailableInstances{}, nil
}
