package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cloudwatchType "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"log"
	"strconv"
)

func (p *AWSCloud) DescribeRegions(ctx context.Context, param types.DescribeRegionsRequest) (types.DescribeRegions, error) {
	allRegion := false
	input := &ec2.DescribeRegionsInput{
		AllRegions: &allRegion,
	}
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
	}
	if response.Regions != nil {
		result := make([]types.ItemRegion, 0, len(response.Regions))
		for _, regin := range response.Regions {
			newRegion := types.ItemRegion{
				RegionId:  aws.StringValue(regin.RegionName),
				LocalName: _regionLocalName[aws.StringValue(regin.RegionName)],
			}
			result = append(result, newRegion)
		}
		return types.DescribeRegions{
			List: result,
		}, err
	}
	return types.DescribeRegions{}, err
}

func (p *AWSCloud) DescribeInstances(ctx context.Context, param types.DescribeInstancesRequest) (types.DescribeInstances, error) {
	awsInstances := make([]types.ItemDescribeInstance, 0)
	input := &ec2.DescribeInstancesInput{
		InstanceIds: param.InstanceIds,
	}
	output, err := p.ec2Client.DescribeInstances(ctx, input)
	log.Println(output.Reservations)
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
		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				region := aws.StringValue(instance.Placement.AvailabilityZone)
				newInstance := types.ItemDescribeInstance{
					InstanceId:   aws.StringValue(instance.InstanceId),
					InstanceName: aws.StringValue(instance.Tags[0].Value),
					RegionId:     aws.StringValue(instance.Placement.AvailabilityZone),
					RegionName:   _regionLocalName[region[0:len(region)-1]],
					/*HostName:           "",
					SubscriptionType:   "",
					InternetChargeType: "",*/
					PublicIpAddress: []string{aws.StringValue(instance.PublicIpAddress)},
					InnerIpAddress:  []string{aws.StringValue(instance.PrivateIpAddress)},
				}
				awsInstances = append(awsInstances, newInstance)
			}
		}
		return types.DescribeInstances{
			TotalCount: len(awsInstances),
			List:       awsInstances,
		}, err
	}
	return types.DescribeInstances{}, err
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
