package datareader

import (
	"context"
	"log"
	"time"

	"github.com/galaxy-future/costpilot/internal/data"
	"github.com/galaxy-future/costpilot/internal/providers"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/galaxy-future/costpilot/tools"
	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
)

type CostDataReader struct {
	_provider providers.Provider
}

func NewCostDataReader(p providers.Provider) *CostDataReader {
	return &CostDataReader{
		_provider: p,
	}
}

// GetDailyCost
// date 2022-09-06 | isGroupByProduct true/false
func (s *CostDataReader) GetDailyCost(ctx context.Context, day string, isGroupByProduct bool) (data.DailyBilling, error) {
	if !tools.IsValidDayDate(day) {
		log.Printf("W! invalid day[%v]\n", day)
		return data.DailyBilling{}, nil
	}
	params := types.QueryAccountBillRequest{
		BillingCycle:     tools.Date2Month(day),
		BillingDate:      day,
		IsGroupByProduct: isGroupByProduct,
		Granularity:      types.Daily,
	}
	resp, err := s._provider.QueryAccountBill(ctx, params)
	if err != nil {
		log.Printf("E! [D] QueryAccountBill error[%v]\n", err)
		return data.DailyBilling{}, err
	}
	result := data.DailyBilling{
		Day:             day,
		ProductsBilling: make(map[string]data.ProductBilling, 0),
	}
	for _, d := range resp.Items.Item {
		result.TotalAmount = tools.Float64Add(result.TotalAmount, cast.ToFloat64(d.PretaxAmount))
		item := data.ItemInProductBilling{
			PipCode:          d.PipCode.String(),
			ProductName:      d.ProductName,
			PretaxAmount:     d.PretaxAmount,
			SubscriptionType: d.SubscriptionType,
			Currency:         d.Currency,
		}
		productCost, ok := result.ProductsBilling[item.PipCode]
		if !ok { // first item
			productCost = data.ProductBilling{
				ProductName: item.ProductName,
				TotalAmount: item.PretaxAmount,
				Items:       []data.ItemInProductBilling{item},
			}
			result.ProductsBilling[item.PipCode] = productCost
			continue
		}
		productCost.TotalAmount = tools.Float64Add(productCost.TotalAmount, item.PretaxAmount)
		productCost.Items = append(productCost.Items, item)
		result.ProductsBilling[item.PipCode] = productCost
	}
	log.Printf("I! GetDailyCost[%s] done \n", day)
	return result, nil
}

// GetDaysCost
// days ["2022-10-01","2022-10-02",]
func (s *CostDataReader) GetDaysCost(ctx context.Context, isGroupByProduct bool, days ...string) ([]data.DailyBilling, error) {
	result := []data.DailyBilling{}
	if len(days) == 0 {
		return result, nil
	}
	sg, ctx := errgroup.WithContext(ctx)
	rCnt := 0
	for _, day := range days {
		d := day
		sg.Go(func() error {
			select {
			case <-ctx.Done():
				log.Printf("I! Canceled GetDaysCost[%s]\n", d)
				return nil
			default:
				res, err := s.GetDailyCost(ctx, d, isGroupByProduct)
				if err != nil {
					return err
				}
				result = append(result, res)
				return nil
			}
		})
		rCnt++
		if rCnt%10 == 0 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	if err := sg.Wait(); err != nil {
		return nil, err
	}
	log.Printf("I! GetDaysCost[%v] done \n", days)
	return result, nil
}

// GetMonthlyCost
// month 2022-09
func (s *CostDataReader) GetMonthlyCost(ctx context.Context, month string, isGroupByProduct bool) (data.MonthlyBilling, error) {
	if !tools.IsValidMonthDate(month) {
		log.Printf("W! invalid month[%v]\n", month)
		return data.MonthlyBilling{}, nil
	}
	params := types.QueryAccountBillRequest{
		BillingCycle:     month,
		IsGroupByProduct: isGroupByProduct,
		Granularity:      types.Monthly,
	}
	resp, err := s._provider.QueryAccountBill(ctx, params)
	if err != nil {
		log.Printf("E! [M] QueryAccountBill error[%v]\n", err)
		return data.MonthlyBilling{}, err
	}
	result := data.MonthlyBilling{
		Month:           month,
		ProductsBilling: make(map[string]data.ProductBilling, 0),
	}
	for _, d := range resp.Items.Item {
		result.TotalAmount = tools.Float64Add(result.TotalAmount, cast.ToFloat64(d.PretaxAmount))
		item := data.ItemInProductBilling{
			PipCode:          d.PipCode.String(),
			ProductName:      d.ProductName,
			PretaxAmount:     d.PretaxAmount,
			SubscriptionType: d.SubscriptionType,
			Currency:         d.Currency,
		}
		productCost, ok := result.ProductsBilling[item.PipCode]
		if !ok { // first item
			productCost = data.ProductBilling{
				ProductName: item.ProductName,
				TotalAmount: item.PretaxAmount,
				Items:       []data.ItemInProductBilling{item},
			}
			result.ProductsBilling[item.PipCode] = productCost
			continue
		}
		productCost.TotalAmount = tools.Float64Add(productCost.TotalAmount, item.PretaxAmount)
		productCost.Items = append(productCost.Items, item)
		result.ProductsBilling[item.PipCode] = productCost
	}
	log.Printf("I! GetMonthlyCost[%v] done\n", month)
	return result, nil
}

// GetMonthsCost
func (s *CostDataReader) GetMonthsCost(ctx context.Context, isGroupByProduct bool, months ...string) ([]data.MonthlyBilling, error) {
	result := []data.MonthlyBilling{}
	if len(months) == 0 {
		return result, nil
	}
	sg, ctx := errgroup.WithContext(ctx)
	rCnt := 0
	for _, month := range months {
		m := month
		sg.Go(func() error {
			select {
			case <-ctx.Done():
				log.Printf("I! Canceled GetMonthsCost[%s]\n", m)
				return nil
			default:
				res, err := s.GetMonthlyCost(ctx, m, isGroupByProduct)
				if err != nil {
					return err
				}
				result = append(result, res)
				return nil
			}
		})
		rCnt++
		if rCnt%10 == 0 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	if err := sg.Wait(); err != nil {
		return nil, err
	}
	log.Printf("get GetMonthsCost[%v] done \n", months)
	return result, nil
}
