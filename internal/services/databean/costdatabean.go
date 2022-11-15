package databean

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/galaxy-future/costpilot/internal/services/datareader"

	"github.com/galaxy-future/costpilot/internal/providers"
	"github.com/galaxy-future/costpilot/internal/types"
	"github.com/galaxy-future/costpilot/tools"
)

type CostDataBean struct {
	billingDate   tools.BillingDate
	daysBilling   sync.Map
	monthsBilling sync.Map

	bp       *tools.BillingDatePilot
	provider providers.Provider

	pipeLineFunc []func(context.Context) error
}

func NewCostDataBean(a types.CloudAccount, t time.Time) *CostDataBean {
	s := &CostDataBean{
		billingDate: tools.BillingDate{},
		bp:          tools.NewBillDatePilot().SetNowT(t),
	}
	s.initProvider(a)

	return s
}

// initProvider
func (s *CostDataBean) initProvider(a types.CloudAccount) *CostDataBean {
	var err error
	s.provider, err = providers.GetProvider(a.Provider, a.AK, a.SK, a.RegionID)
	if err != nil {
		log.Printf("E! init provider failed: %v\n", err)
	}
	return s
}

// GetBillingMap
func (s *CostDataBean) GetBillingMap() (*sync.Map, *sync.Map) {
	return &s.monthsBilling, &s.daysBilling
}

// getRecent15DaysBilling today is not included
func (s *CostDataBean) getRecent15DaysBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentXDaysBillingDate(15)
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	// log.Printf("I! getRecent15DaysBilling done")
	return nil
}

// getPreviousYearRecent15DaysBilling
func (s *CostDataBean) getPreviousYearRecent15DaysBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentXDaysBillingDate(15)
	days := billingDate.Days
	lastYearDays := s.bp.GetTargetYearData(days, -1)
	billingDate.Days = lastYearDays
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	// log.Printf("I! getPreviousYearRecent15DaysBilling done")
	return nil
}

// getLast12MonthsBilling
// current month is included, but data of today is not included
func (s *CostDataBean) getRecent24MonthsBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentXMonthsBillingDate(24)
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	// log.Printf("I! getLast12MonthsBilling done")
	return nil
}

// getRecentYearMonthsBilling
// if today is 01-01, current year is last year
func (s *CostDataBean) getRecentYearMonthsBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentYearBillingDate()
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	// log.Printf("I! getRecentYearMonthsBilling done")
	return nil
}

// getPreviousYearMonthsBilling
// if today is 01-01, last year is before last year
func (s *CostDataBean) getPreviousYearMonthsBilling(ctx context.Context) error {
	billingDate := s.bp.GetPreviousYearBillingDate()
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	// log.Printf("I! getPreviousYearMonthsBilling done")
	return nil
}

// getRecentDayBilling
func (s *CostDataBean) getRecentDayBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentDayBillingDate()
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	// log.Printf("I! getRecentDayBilling done")
	return nil
}

// getPreviousDayDayBilling
func (s *CostDataBean) getPreviousDayDayBilling(ctx context.Context) error {
	billingDate := s.bp.GetPreviousDayBillingDate()
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	// log.Printf("I! getPreviousDayDayBilling done")
	return nil
}

// getRecentDayBillingWithProduct
func (s *CostDataBean) getRecentDayBillingWithProduct(ctx context.Context) error {
	billingDate := s.bp.GetRecentDayBillingDate()
	day := billingDate.Days[0]
	costDataReader := datareader.NewCostDataReader(s.provider)
	dayBilling, err := costDataReader.GetDailyCost(ctx, day, true)
	if err != nil {
		return err
	}
	s.daysBilling.Store(dayBilling.Day, dayBilling) // cover old data
	log.Printf("I! getRecentDayBillingWithProduct done")
	return nil
}

// getRecentMonthBillingWithProduct
func (s *CostDataBean) getRecentMonthBillingWithProduct(ctx context.Context) error {
	monthBillingDate := s.bp.GetRecentMonthBillingDate(true)
	costDataReader := datareader.NewCostDataReader(s.provider)
	if len(monthBillingDate.Months) != 0 {
		monthsBilling, err := costDataReader.GetMonthsCost(ctx, true, monthBillingDate.Months...)
		if err != nil {
			return err
		}
		for _, v := range monthsBilling {
			s.monthsBilling.Store(v.Month, v) // cover old data
		}

	}

	if len(monthBillingDate.Days) != 0 {
		daysBilling, err := costDataReader.GetDaysCost(ctx, true, monthBillingDate.Days...)
		if err != nil {
			return err
		}
		for _, v := range daysBilling {
			s.daysBilling.Store(v.Day, v)
		}
	}

	log.Printf("I! getRecentMonthBillingWithProduct done")
	return nil
}

// getRecentQuarterBilling
func (s *CostDataBean) getRecentQuarterBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentQuarterBillingDate(true)
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	// log.Printf("I! getRecentQuarterBilling done")
	return nil
}

// getPreviousQuarterBilling
func (s *CostDataBean) getPreviousQuarterBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentQuarterBillingDate(true)
	billingDate = s.bp.ConvBillingDate2PreviousQuarter(billingDate)
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	// log.Printf("I! getPreviousQuarterBilling done")
	return nil
}

// getPreviousMouthBilling
func (s *CostDataBean) getPreviousMouthBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentMonthBillingDate(true)
	billingDate = s.bp.ConvBillingDate2PreviousMonth(billingDate)
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	// log.Printf("I! getPreviousMouthBilling done")
	return nil
}

// AddBillingDate
func (s *CostDataBean) AddBillingDate(ctx context.Context, billingDate tools.BillingDate) error {
	s.billingDate.Months = tools.Union(s.billingDate.Months, billingDate.Months)
	s.billingDate.Days = tools.Union(s.billingDate.Days, billingDate.Days)
	return nil
}

// FillBillings
func (s *CostDataBean) FillBillings(ctx context.Context) error {
	b := s.billingDate
	costDataReader := datareader.NewCostDataReader(s.provider)
	var months, days []string
	for _, v := range b.Months {
		if _, ok := s.monthsBilling.Load(v); !ok { // skip if key exist
			months = append(months, v)
		}
	}
	for _, v := range b.Days {
		if _, ok := s.daysBilling.Load(v); !ok {
			days = append(days, v)
		}
	}
	sort.Slice(months, func(i, j int) bool {
		return months[i] < months[j]
	})
	sort.Slice(days, func(i, j int) bool {
		return days[i] < days[j]
	})
	monthsBilling, err := costDataReader.GetMonthsCost(ctx, false, months...)
	if err != nil {
		return err
	}
	daysBilling, err := costDataReader.GetDaysCost(ctx, false, days...)
	if err != nil {
		return err
	}
	for _, v := range monthsBilling {
		s.monthsBilling.LoadOrStore(v.Month, v)
	}
	for _, v := range daysBilling {
		s.daysBilling.LoadOrStore(v.Day, v)
	}

	return nil
}

// GetCostAnalysisPipeLine
func (s *CostDataBean) GetCostAnalysisPipeLine() []func(context.Context) error {
	return []func(context.Context) error{
		s.getRecent24MonthsBilling,
		s.getRecentYearMonthsBilling,
		s.getPreviousYearMonthsBilling,
		s.getRecent15DaysBilling,
		s.getPreviousYearRecent15DaysBilling,
		s.getRecentQuarterBilling,
		s.getPreviousQuarterBilling,
		s.getPreviousMouthBilling,
		s.getRecentDayBilling,
		s.getPreviousDayDayBilling,
		s.FillBillings,
		//
		s.getRecentDayBillingWithProduct,
		s.getRecentMonthBillingWithProduct,
	}
}

// RunPipeline
func (s *CostDataBean) RunPipeline(ctx context.Context) error {
	var err error
	for _, f := range s.GetCostAnalysisPipeLine() {
		err = f(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
