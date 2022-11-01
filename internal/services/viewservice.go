package services

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/galayx-future/costpilot/internal/providers"
	"github.com/galayx-future/costpilot/internal/types"
	"github.com/galayx-future/costpilot/tools"
)

type ViewService struct {
	billingDate   tools.BillingDate
	daysBilling   sync.Map
	monthsBilling sync.Map

	bp       *tools.BillingDatePilot
	provider providers.Provider

	pipeLineFunc []func(context.Context) error
}

func NewViewService(a types.CloudAccount, t time.Time) *ViewService {
	s := &ViewService{
		billingDate: tools.BillingDate{},
		bp:          tools.NewBillDatePilot().SetNowT(t),
	}
	s.initProvider(a)

	return s
}

// initProvider
func (s *ViewService) initProvider(a types.CloudAccount) *ViewService {
	var err error
	s.provider, err = providers.GetProvider(a.Provider, a.AK, a.SK, a.RegionID)
	if err != nil {
		log.Printf("E! init provider failed: %v\n", err)
	}
	return s
}

// GetBillingMap
func (s *ViewService) GetBillingMap() (*sync.Map, *sync.Map) {
	return &s.monthsBilling, &s.daysBilling
}

// getRecent15DaysBilling today is not included
func (s *ViewService) getRecent15DaysBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentXDaysBillingDate(15)
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	//log.Printf("I! getRecent15DaysBilling done")
	return nil
}

// getPreviousYearRecent15DaysBilling
func (s *ViewService) getPreviousYearRecent15DaysBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentXDaysBillingDate(15)
	days := billingDate.Days
	lastYearDays := s.bp.GetTargetYearData(days, -1)
	billingDate.Days = lastYearDays
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	//log.Printf("I! getPreviousYearRecent15DaysBilling done")
	return nil
}

// getLast12MonthsBilling
// current month is included, but data of today is not included
func (s *ViewService) getRecent24MonthsBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentXMonthsBillingDate(24)
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	//log.Printf("I! getLast12MonthsBilling done")
	return nil
}

// getRecentYearMonthsBilling
// if today is 01-01, current year is last year
func (s *ViewService) getRecentYearMonthsBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentYearBillingDate()
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	//log.Printf("I! getRecentYearMonthsBilling done")
	return nil
}

// getPreviousYearMonthsBilling
// if today is 01-01, last year is before last year
func (s *ViewService) getPreviousYearMonthsBilling(ctx context.Context) error {
	billingDate := s.bp.GetPreviousYearBillingDate()
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	//log.Printf("I! getPreviousYearMonthsBilling done")
	return nil
}

// getRecentDayBilling
func (s *ViewService) getRecentDayBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentDayBillingDate()
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	//log.Printf("I! getRecentDayBilling done")
	return nil
}

// getPreviousDayDayBilling
func (s *ViewService) getPreviousDayDayBilling(ctx context.Context) error {
	billingDate := s.bp.GetPreviousDayBillingDate()
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	//log.Printf("I! getPreviousDayDayBilling done")
	return nil
}

// getRecentDayBillingWithProduct
func (s *ViewService) getRecentDayBillingWithProduct(ctx context.Context) error {
	billingDate := s.bp.GetRecentDayBillingDate()
	day := billingDate.Days[0]
	costSvc := NewCostService(s.provider)
	dayBilling, err := costSvc.GetDailyCost(ctx, day, true)
	if err != nil {
		return err
	}
	s.daysBilling.Store(dayBilling.Day, dayBilling) //cover old data
	log.Printf("I! getRecentDayBillingWithProduct done")
	return nil
}

// getRecentMonthBillingWithProduct
func (s *ViewService) getRecentMonthBillingWithProduct(ctx context.Context) error {
	monthBillingDate := s.bp.GetRecentMonthBillingDate(true)
	costSvc := NewCostService(s.provider)
	if len(monthBillingDate.Months) != 0 {
		monthsBilling, err := costSvc.GetMonthsCost(ctx, true, monthBillingDate.Months...)
		if err != nil {
			return err
		}
		for _, v := range monthsBilling {
			s.monthsBilling.Store(v.Month, v) //cover old data
		}

	}

	if len(monthBillingDate.Days) != 0 {
		daysBilling, err := costSvc.GetDaysCost(ctx, true, monthBillingDate.Days...)
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
func (s *ViewService) getRecentQuarterBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentQuarterBillingDate(true)
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	//log.Printf("I! getRecentQuarterBilling done")
	return nil
}

// getPreviousQuarterBilling
func (s *ViewService) getPreviousQuarterBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentQuarterBillingDate(true)
	billingDate = s.bp.ConvBillingDate2PreviousQuarter(billingDate)
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	//log.Printf("I! getPreviousQuarterBilling done")
	return nil
}

// getPreviousMouthBilling
func (s *ViewService) getPreviousMouthBilling(ctx context.Context) error {
	billingDate := s.bp.GetRecentMonthBillingDate(true)
	billingDate = s.bp.ConvBillingDate2PreviousMonth(billingDate)
	if err := s.AddBillingDate(ctx, billingDate); err != nil {
		return err
	}
	//log.Printf("I! getPreviousMouthBilling done")
	return nil
}

// AddBillingDate
func (s *ViewService) AddBillingDate(ctx context.Context, billingDate tools.BillingDate) error {
	s.billingDate.Months = tools.Union(s.billingDate.Months, billingDate.Months)
	s.billingDate.Days = tools.Union(s.billingDate.Days, billingDate.Days)
	return nil
}

// FillBillings
func (s *ViewService) FillBillings(ctx context.Context) error {
	b := s.billingDate
	costSvc := NewCostService(s.provider)
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
	monthsBilling, err := costSvc.GetMonthsCost(ctx, false, months...)
	if err != nil {
		return err
	}
	daysBilling, err := costSvc.GetDaysCost(ctx, false, days...)
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
func (s *ViewService) GetCostAnalysisPipeLine() []func(context.Context) error {
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
func (s *ViewService) RunPipeline(ctx context.Context) error {
	var err error
	for _, f := range s.GetCostAnalysisPipeLine() {
		err = f(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
