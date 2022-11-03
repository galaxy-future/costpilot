package tools

import (
	"errors"
	"fmt"
	"math"

	"github.com/galaxy-future/costpilot/internal/data"

	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

func Float64Add(a ...float64) float64 {
	if len(a) == 0 {
		return 0
	}
	if len(a) == 1 {
		return a[0]
	}
	var result decimal.Decimal
	for _, v := range a {
		result = result.Add(decimal.NewFromFloat(v))
	}
	return result.InexactFloat64()
}

func RatioString(s1, s2 string) string {
	m := cast.ToFloat64(s1)
	n := cast.ToFloat64(s2)
	if decimal.NewFromFloat(m).Equal(decimal.NewFromFloat(0)) {
		return "--"
	}
	return fmt.Sprintf("%0.2f", 100*decimal.NewFromFloat(n).Sub(decimal.NewFromFloat(m)).DivRound(decimal.NewFromFloat(math.Abs(m)), 4).InexactFloat64())
}

func AddDailyBilling(x, y data.DailyBilling) data.DailyBilling {
	var ret data.DailyBilling
	ret.TotalAmount = x.TotalAmount + y.TotalAmount
	ret.Day = y.Day
	ret.ProductsBilling = make(map[string]data.ProductBilling)
	for pipcode, bill := range x.ProductsBilling {
		ret.ProductsBilling[pipcode] = bill
	}
	for pipcode, bill := range y.ProductsBilling {
		if val, ok := ret.ProductsBilling[pipcode]; ok {
			t, _ := AddProductBilling(val, bill)
			ret.ProductsBilling[pipcode] = t
		} else {
			ret.ProductsBilling[pipcode] = bill
		}
	}
	return ret
}

func AddMonthlyBilling(x, y data.MonthlyBilling) data.MonthlyBilling {
	var ret data.MonthlyBilling
	ret.TotalAmount = x.TotalAmount + y.TotalAmount
	ret.Month = y.Month
	ret.ProductsBilling = make(map[string]data.ProductBilling)
	for pipcode, bill := range x.ProductsBilling {
		ret.ProductsBilling[pipcode] = bill
	}
	for pipcode, bill := range y.ProductsBilling {
		if val, ok := ret.ProductsBilling[pipcode]; ok {
			t, _ := AddProductBilling(val, bill)
			ret.ProductsBilling[pipcode] = t
		} else {
			ret.ProductsBilling[pipcode] = bill
		}
	}
	return ret
}
func AddProductBilling(x, y data.ProductBilling) (data.ProductBilling, error) {
	var ret data.ProductBilling
	if x.ProductName != y.ProductName {
		return ret, errors.New("two products are not the same")
	}
	ret.Items = make([]data.ItemInProductBilling, len(x.Items))
	copy(ret.Items, x.Items)
	ret.ProductName = x.ProductName
	ret.TotalAmount = x.TotalAmount + y.TotalAmount
	for _, itemy := range y.Items {
		exist := false
		for k, item := range ret.Items {
			if itemy.SubscriptionType == item.SubscriptionType {
				ret.Items[k].PretaxAmount += itemy.PretaxAmount
				exist = true
			}
		}
		if !exist {
			ret.Items = append(ret.Items, itemy)
		}

	}
	return ret, nil
}
