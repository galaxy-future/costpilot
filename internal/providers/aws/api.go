package aws

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/alibabacloud-go/tea/tea"
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
	client *costexplorer.Client
}

func New(AK, SK, regionId string) (*AWSCloud, error) {

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(regionId), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(AK, SK, "")))
	if err != nil {
		return nil, err
	}
	return &AWSCloud{client: costexplorer.NewFromConfig(cfg)}, nil
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

func (p *AWSCloud) DescribeMetricList(ctx context.Context, param types.DescribeMetricListRequest) (types.DescribeMetricList, error) {
	// TODO implement me
	panic("implement me")
}

func (p *AWSCloud) DescribeInstanceAttribute(ctx context.Context, param types.DescribeInstanceAttributeRequest) (types.DescribeInstanceAttribute, error) {
	// TODO implement me
	panic("implement me")
}

func (p *AWSCloud) DescribeRegions(ctx context.Context, param types.DescribeRegionsRequest) (types.DescribeRegions, error) {
	// TODO implement me
	panic("implement me")
}

func (p *AWSCloud) DescribeInstanceBill(ctx context.Context, param types.DescribeInstanceBillRequest, isAll bool) (types.DescribeInstanceBill, error) {

	// TODO implement me
	panic("implement me")
}

func (p *AWSCloud) QueryAvailableInstances(ctx context.Context, param types.QueryAvailableInstancesRequest) (types.QueryAvailableInstances, error) {
	// TODO implement me
	panic("implement me")
}
