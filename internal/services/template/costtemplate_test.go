package template

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/data"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/galaxy-future/costpilot/tools"
)

var (
	s             CostTemplate
	daysBilling   sync.Map
	monthsBilling sync.Map
)

func TestMain(m *testing.M) {
	for month := 1; month <= 12; month++ {
		switch month {
		case 1, 3, 5, 7, 8, 10, 12:
			for day := 1; day <= 31; day++ {
				fillDailyBill(fmt.Sprintf("2022-%02d-%02d", month, day), float64(2*month*day), float64(3*month*day), float64(4*month*day))
				fillDailyBill(fmt.Sprintf("2021-%02d-%02d", month, day), float64(month*day), float64(2*month*day), float64(3*month*day))
				fillDailyBill(fmt.Sprintf("2020-%02d-%02d", month, day), float64(2*month*day), float64(4*month*day), float64(6*month*day))
			}
			fillMonthlyBill(fmt.Sprintf("2022-%02d", month), float64(32*31*month), float64(32*31*3/2*month), float64(2*month*32*31))
			fillMonthlyBill(fmt.Sprintf("2021-%02d", month), float64(32*31/2*month), float64(32*31*month), float64(3/2*month*32*31))
			fillMonthlyBill(fmt.Sprintf("2020-%02d", month), float64(32*31*month), float64(2*32*31*month), float64(3*month*32*31))
		case 2:
			for day := 1; day <= 28; day++ {
				fillDailyBill(fmt.Sprintf("2022-%02d-%02d", month, day), float64(2*month*day), float64(3*month*day), float64(4*month*day))
				fillDailyBill(fmt.Sprintf("2021-%02d-%02d", month, day), float64(month*day), float64(2*month*day), float64(3*month*day))
				fillDailyBill(fmt.Sprintf("2020-%02d-%02d", month, day), float64(2*month*day), float64(4*month*day), float64(6*month*day))
			}
			fillMonthlyBill(fmt.Sprintf("2022-%02d", month), float64(29*28/2*2*month), float64(29*28/2*3*month), float64(29*28/2*4*month))
			fillMonthlyBill(fmt.Sprintf("2021-%02d", month), float64(29*28/2*month), float64(29*28/2*2*month), float64(29*28/2*3*month))
			fillMonthlyBill(fmt.Sprintf("2020-%02d", month), float64(29*28*month), float64(29*28*2*month), float64(29*28*3*month))
		default:
			for day := 1; day <= 30; day++ {
				fillDailyBill(fmt.Sprintf("2022-%02d-%02d", month, day), float64(2*month*day), float64(3*month*day), float64(4*month*day))
				fillDailyBill(fmt.Sprintf("2021-%02d-%02d", month, day), float64(month*day), float64(2*month*day), float64(3*month*day))
				fillDailyBill(fmt.Sprintf("2020-%02d-%02d", month, day), float64(2*month*day), float64(4*month*day), float64(6*month*day))
			}
			fillMonthlyBill(fmt.Sprintf("2022-%02d", month), float64(30*31/2*2*month), float64(30*31/2*3*month), float64(30*31/2*4*month))
			fillMonthlyBill(fmt.Sprintf("2021-%02d", month), float64(30*31/2*month), float64(30*31/2*2*month), float64(30*31/2*3*month))
			fillMonthlyBill(fmt.Sprintf("2020-%02d", month), float64(30*31*month), float64(30*31*2*month), float64(30*31*3*month))
		}
	}

	s = CostTemplate{
		DaysBilling:   &daysBilling,
		MonthsBilling: &monthsBilling,
		bp:            tools.NewBillDatePilot(),
	}
	os.Exit(m.Run())
}
func TestCombineBilling(t *testing.T) {
	var monthsBillingList sync.Map
	monthsBillingList.Store("2022-02", data.MonthlyBilling{
		Month: "2022-02",
		ProductsBilling: map[string]data.ProductBilling{
			types.ECS.String(): {
				ProductName: types.ECS.String(),
				TotalAmount: 10,
				Items: []data.ItemInProductBilling{
					{
						PipCode:          types.ECS.String(),
						ProductName:      types.ECS.String(),
						PretaxAmount:     6,
						SubscriptionType: cloud.PrePaid,
					},
					{
						PipCode:          types.ECS.String(),
						ProductName:      types.ECS.String(),
						PretaxAmount:     4,
						SubscriptionType: cloud.PostPaid,
					},
				},
			},
			types.S3.String(): {
				ProductName: types.ECS.String(),
				TotalAmount: 20,
				Items:       nil,
			},
		},
		TotalAmount: 30,
	})
	tests := []struct {
		monthsBillingList []*sync.Map
		daysBillingList   []*sync.Map
	}{
		{
			monthsBillingList: []*sync.Map{&monthsBillingList},
			daysBillingList:   nil,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			s.CombineBilling(context.Background(), tt.monthsBillingList, tt.daysBillingList)
			s.MonthsBilling.Range(func(key, value interface{}) bool {
				fmt.Println(key, "\n", value.(data.MonthlyBilling).ProductsBilling, "\n", value.(data.MonthlyBilling).TotalAmount)
				return true
			})
		})
	}
}
func TestGetStatisics(t *testing.T) {
	time, _ := time.Parse("2006-01-02", "2022-03-01")
	s.bp.SetNowT(time)
	fmt.Println(s.getStatistics())
}
func TestGetItemInSeries(t *testing.T) {
	time, _ := time.Parse("2006-01-02", "2022-03-11")
	s.bp.SetNowT(time)
	fmt.Println(s.amountInLastXMonths(14, true))
	fmt.Println(s.amountInLastXMonths(14, false))
}
func TestFormatStatistics(t *testing.T) {
	time, _ := time.Parse("2006-01-02", "2022-03-11")
	s.bp.SetNowT(time)
	temp, _ := s.FormatMonthStatistics(context.Background())
	fmt.Printf("%#v", temp)
}
func fillDailyBill(day string, ecsTotal, s3Total, diskTotal float64) {
	d := data.DailyBilling{
		Day:             day,
		ProductsBilling: make(map[string]data.ProductBilling),
		TotalAmount:     ecsTotal + s3Total + diskTotal,
	}
	if ecsTotal > 0 {
		d.ProductsBilling[types.ECS.String()] = data.ProductBilling{
			ProductName: types.ECS.String(),
			TotalAmount: ecsTotal,
			Items: []data.ItemInProductBilling{
				{
					PipCode:          types.ECS.String(),
					ProductName:      types.ECS.String(),
					PretaxAmount:     ecsTotal * 0.6,
					SubscriptionType: cloud.PrePaid,
				},
				{
					PipCode:          types.ECS.String(),
					ProductName:      types.ECS.String(),
					PretaxAmount:     ecsTotal * 0.4,
					SubscriptionType: cloud.PostPaid,
				},
			},
		}
	}
	if s3Total > 0 {
		d.ProductsBilling[types.S3.String()] = data.ProductBilling{
			ProductName: types.S3.String(),
			TotalAmount: s3Total,
			Items:       nil,
		}
	}
	if diskTotal > 0 {
		d.ProductsBilling[types.DISK.String()] = data.ProductBilling{
			ProductName: types.DISK.String(),
			TotalAmount: diskTotal,
			Items:       nil,
		}
	}
	daysBilling.Store(day, d)
}
func fillMonthlyBill(month string, ecsTotal, s3Total, diskTotal float64) {
	m := data.MonthlyBilling{
		Month:           month,
		ProductsBilling: make(map[string]data.ProductBilling),
		TotalAmount:     ecsTotal + s3Total + diskTotal,
	}
	if ecsTotal > 0 {
		m.ProductsBilling[types.ECS.String()] = data.ProductBilling{
			ProductName: types.ECS.String(),
			TotalAmount: ecsTotal,
			Items: []data.ItemInProductBilling{
				{
					PipCode:          types.ECS.String(),
					ProductName:      types.ECS.String(),
					PretaxAmount:     ecsTotal * 0.6,
					SubscriptionType: cloud.PrePaid,
				},
				{
					PipCode:          types.ECS.String(),
					ProductName:      types.ECS.String(),
					PretaxAmount:     ecsTotal * 0.4,
					SubscriptionType: cloud.PostPaid,
				},
			},
		}
	}
	if s3Total > 0 {
		m.ProductsBilling[types.S3.String()] = data.ProductBilling{
			ProductName: types.S3.String(),
			TotalAmount: s3Total,
			Items:       nil,
		}
	}
	if diskTotal > 0 {
		m.ProductsBilling[types.DISK.String()] = data.ProductBilling{
			ProductName: types.DISK.String(),
			TotalAmount: diskTotal,
			Items:       nil,
		}
	}
	monthsBilling.Store(month, m)
}
