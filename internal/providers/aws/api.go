package aws

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	explorerTypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/galayx-future/costpilot/internal/constants/cloud"
	"github.com/galayx-future/costpilot/internal/providers/types"
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
func (*AWSCloud) ProviderType() string {
	return cloud.AWSCloud
}

// QueryAccountBill
func (p *AWSCloud) QueryAccountBill(ctx context.Context, param types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error) {
	const dateFormat string = "2006-01-02"
	var start, end time.Time
	var err error
	var billItems []types.AccountBillItem
	if param.Granularity == types.Monthly {
		start, err = time.Parse(dateFormat, param.BillingCycle+"-01")
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
		end = start.AddDate(0, 1, 0)
		if err != nil {
			return types.DataInQueryAccountBill{}, err
		}
	} else if param.Granularity == types.Daily {
		start, err = time.Parse(dateFormat, param.BillingDate)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
		end = start.AddDate(0, 0, 1)
		if err != nil {
			return types.DataInQueryAccountBill{}, err
		}
	}
	input := &costexplorer.GetCostAndUsageInput{
		Granularity: convGranularity(param.Granularity),
		TimePeriod: &explorerTypes.DateInterval{
			End:   aws.String(end.Format(dateFormat)),
			Start: aws.String(start.Format(dateFormat)),
		},
		Metrics: []string{"BlendedCost"},
	}
	if param.IsGroupByProduct {
		input.GroupBy = []explorerTypes.GroupDefinition{
			{
				Key:  aws.String(string(explorerTypes.DimensionService)),
				Type: explorerTypes.GroupDefinitionTypeDimension,
			},
		}
	}
	for {
		output, err := p.client.GetCostAndUsage(ctx, input)
		if err != nil {
			return types.DataInQueryAccountBill{}, err
		}
		newbillItems, err := convAccountBillItems(output, param)
		if err != nil {
			return types.DataInQueryAccountBill{}, err
		}
		billItems = append(billItems, newbillItems...)
		if output != nil && output.NextPageToken != nil {
			input.NextPageToken = output.NextPageToken
		} else {
			break
		}
	}
	result := types.DataInQueryAccountBill{
		BillingCycle: param.BillingCycle,
		AccountID:    "undefined", // no such field in the AWS API response
		TotalCount:   len(billItems),
		AccountName:  "undefined", // no such field in the AWS API response
		Items: types.ItemsInQueryAccountBill{
			Item: billItems,
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
func convAccountBillItems(output *costexplorer.GetCostAndUsageOutput, param types.QueryAccountBillRequest) ([]types.AccountBillItem, error) {
	var result []types.AccountBillItem
	var err error
	resultBytime := output.ResultsByTime[0]
	if param.IsGroupByProduct {
		result = make([]types.AccountBillItem, 0, len(resultBytime.Groups))
		for _, group := range resultBytime.Groups {
			newItem := types.AccountBillItem{
				SubscriptionType: "undefined", // no such field in the AWS API response
				PipCode:          "undefined", // no such field in the AWS API response
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
			SubscriptionType: "undefined", // no such field in the AWS API response
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

// convert a string pointer to float64 with 2 decimal places
func convAmount(amountPtr *string) (float64, error) {
	if amountPtr == nil {
		return 0, errors.New("amountPtr is nil")
	}
	amountStr := aws.StringValue(amountPtr)
	dotIdx := -1
	for i := 0; i < len(amountStr); i++ {
		if amountStr[i] == '.' {
			dotIdx = i
			break
		}
	}
	if dotIdx != -1 && dotIdx+3 <= len(amountStr) {
		return strconv.ParseFloat(amountStr[:dotIdx+3], 64)
	}
	return strconv.ParseFloat(amountStr, 64)
}
