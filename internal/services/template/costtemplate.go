package template

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/galaxy-future/costpilot/internal/constants"
	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/data"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/galaxy-future/costpilot/internal/template"
	"github.com/galaxy-future/costpilot/tools"
	"github.com/spf13/cast"
)

type CostTemplate struct {
	DaysBilling   *sync.Map // key : day , val : data.DailyBilling
	MonthsBilling *sync.Map // key : month , val : data.MonthlyBilling

	bp           *tools.BillingDatePilot
	analysisData template.AnalysisData

	provider cloud.Provider // tmp solution for multiple cloud provider TODO delete
}

func NewCostTemplate(monthsBilling, daysBilling *sync.Map, t time.Time) *CostTemplate {
	return &CostTemplate{
		MonthsBilling: monthsBilling,
		DaysBilling:   daysBilling,
		bp:            tools.NewBillDatePilot().SetNowT(t),
		analysisData:  template.AnalysisData{},
	}
}

func (s *CostTemplate) SetProvider(provider cloud.Provider) {
	s.provider = provider
}

// CombineBilling 重新组合并制定 DaysBilling, MonthsBilling
func (s *CostTemplate) CombineBilling(ctx context.Context, monthsBillingList, daysBillingList []*sync.Map) error {
	for _, monthMap := range monthsBillingList {
		if s.MonthsBilling == nil {
			var t sync.Map
			monthMap.Range(func(key, value interface{}) bool {
				t.Store(key, value)
				return true
			})
			s.MonthsBilling = &t
			continue
		}
		s.MonthsBilling.Range(func(key, value interface{}) bool {
			if val, ok := monthMap.Load(key); ok {
				s.MonthsBilling.Store(key, tools.AddMonthlyBilling(val.(data.MonthlyBilling), value.(data.MonthlyBilling)))
			}
			return true
		})
		monthMap.Range(func(key, value interface{}) bool {
			if _, ok := s.MonthsBilling.Load(key); !ok {
				s.MonthsBilling.Store(key, value)
			}
			return true
		})
	}
	for _, dayMap := range daysBillingList {
		if s.DaysBilling == nil {
			var t sync.Map
			dayMap.Range(func(key, value interface{}) bool {
				t.Store(key, value)
				return true
			})
			s.DaysBilling = &t
			continue
		}
		s.DaysBilling.Range(func(key, value interface{}) bool {
			if val, ok := dayMap.Load(key); ok {
				s.DaysBilling.Store(key, tools.AddDailyBilling(val.(data.DailyBilling), value.(data.DailyBilling)))
			}
			return true
		})
		dayMap.Range(func(key, value interface{}) bool {
			if _, ok := s.DaysBilling.Load(key); !ok {
				s.DaysBilling.Store(key, value)
			}
			return true
		})
	}
	log.Printf("I! CombineBilling done")
	return nil
}

// FormatDayStatistics
func (s *CostTemplate) FormatDayStatistics(ctx context.Context) (template.CostAnalysis, error) {
	costAnalysisByDay := template.CostAnalysis{
		ViewType:   "day",
		DataCycle:  "",
		Statistics: nil,
		Ratios:     nil,
		CostTrend:  nil,
	}
	recentDay := s.bp.GetRecentDayBillingDate()

	costAnalysisByDay.DataCycle = recentDay.Days[0] + " 23:59:59"
	// Statistics
	costAnalysisByDay.Statistics = s.getStatistics()
	// Ratios
	costAnalysisByDay.Ratios = []template.ItemInRatios{
		template.ItemInRatios{
			Chart: template.ChartInRatios{
				ID:       "productTypeRatio",
				Title:    "日成本构成比例",
				MidUnit:  s.extractCurrencyUnit(),
				MidValue: "",
				Data:     s.productTypeRatioData(recentDay),
			},
		},
		template.ItemInRatios{
			Chart: template.ChartInRatios{
				ID:       "providerTypeRatio",
				Title:    "日成本云厂商比例",
				MidUnit:  s.extractCurrencyUnit(),
				MidValue: "",
				Data:     s.providerTypeRatioData(recentDay),
			},
		},
		template.ItemInRatios{
			Chart: template.ChartInRatios{
				ID:       "chargeTypeRatio",
				Title:    "云服务器日花费付费类型",
				MidUnit:  s.extractCurrencyUnit(),
				MidValue: "",
				Data:     s.chargeTypeRatioData(recentDay),
			},
		},
	}
	// 单独计算 MidValue
	for i, v := range costAnalysisByDay.Ratios {
		costAnalysisByDay.Ratios[i].Chart.MidValue = s.sumRatiosChartMidValue(v.Chart.Data)
	}
	// CostTrend
	costAnalysisByDay.CostTrend = &template.CostTrend{
		Chart: template.ChartInCostTrend{
			ID:     "costTrend",
			Title:  "成本走势",
			XData:  s.getLast14Days(),
			Series: s.getDayItemInSeries(),
			YTitle: []string{"成本 (" + s.extractCurrencyUnit() + ")", "变化率 (%)"},
			TooltipUnit: template.TooltipUnit{
				Bar:  s.extractCurrencyUnit(),
				Line: "%",
			},
		},
	}

	log.Printf("I! FormatDayStatistics done")
	return costAnalysisByDay, nil
}

func (s *CostTemplate) FormatMonthStatistics(ctx context.Context) (template.CostAnalysis, error) {
	costAnalysisByMonth := template.CostAnalysis{
		ViewType:   "month",
		DataCycle:  "",
		Statistics: nil,
		Ratios:     nil,
		CostTrend:  nil,
	}
	yesterday := s.bp.GetRecentDayBillingDate()
	costAnalysisByMonth.DataCycle = yesterday.Days[0] + " 23:59:59"
	recentMonth := s.bp.GetRecentMonthBillingDate(true)
	costAnalysisByMonth.Statistics = s.getStatistics()
	costAnalysisByMonth.Ratios = []template.ItemInRatios{
		template.ItemInRatios{
			Chart: template.ChartInRatios{
				ID:       "productTypeRatio",
				Title:    "月成本构成比例",
				MidUnit:  s.extractCurrencyUnit(),
				MidValue: "",
				Data:     s.productTypeRatioData(recentMonth),
			},
		},
		template.ItemInRatios{
			Chart: template.ChartInRatios{
				ID:       "providerTypeRatio",
				Title:    "月成本云厂商比例",
				MidUnit:  s.extractCurrencyUnit(),
				MidValue: "",
				Data:     s.providerTypeRatioData(recentMonth),
			},
		},
		template.ItemInRatios{
			Chart: template.ChartInRatios{
				ID:       "chargeTypeRatio",
				Title:    "云服务器月花费付费类型",
				MidUnit:  s.extractCurrencyUnit(),
				MidValue: "",
				Data:     s.chargeTypeRatioData(recentMonth),
			},
		},
	}
	// 单独计算 MidValue
	for i, v := range costAnalysisByMonth.Ratios {
		costAnalysisByMonth.Ratios[i].Chart.MidValue = s.sumRatiosChartMidValue(v.Chart.Data)
	}
	// CostTrend
	costAnalysisByMonth.CostTrend = &template.CostTrend{
		Chart: template.ChartInCostTrend{
			ID:     "costTrend",
			Title:  "成本走势",
			XData:  s.getLast12Months(),
			Series: s.getMonthItemInSeries(),
			YTitle: []string{"成本 (" + s.extractCurrencyUnit() + ")", "变化率 (%)"},
			TooltipUnit: template.TooltipUnit{
				Bar:  s.extractCurrencyUnit(),
				Line: "%",
			},
		},
	}

	log.Printf("I! FormatMonthStatistics done")
	return costAnalysisByMonth, nil
}

func (s *CostTemplate) sumBillingDateAmount(date tools.BillingDate) string {
	var sum float64
	for _, m := range date.Months {
		if val, ok := s.MonthsBilling.Load(m); ok {
			sum += val.(data.MonthlyBilling).TotalAmount
		}
	}
	for _, d := range date.Days {
		if val, ok := s.DaysBilling.Load(d); ok {
			sum += val.(data.DailyBilling).TotalAmount
		}
	}
	return fmt.Sprintf("%.2f", sum)
}

func (s *CostTemplate) productTypeRatioData(date tools.BillingDate) []template.ItemInRatioData {
	ret := make([]template.ItemInRatioData, 0)
	totalMap := make(map[string]float64) // key : pipCode, val : totalAmount
	kM := make(map[string]string)        // key :pipCode productName
	for _, d := range date.Days {
		if val, ok := s.DaysBilling.Load(d); ok {
			for k, v := range val.(data.DailyBilling).ProductsBilling {
				totalMap[k] += v.TotalAmount
				kM[k] = v.ProductName
			}
		}
	}
	for _, m := range date.Months {
		if val, ok := s.MonthsBilling.Load(m); ok {
			for k, v := range val.(data.MonthlyBilling).ProductsBilling {
				totalMap[k] += v.TotalAmount
				kM[k] = v.ProductName
			}
		}
	}
	for k, v := range totalMap {
		/*		name := types.PidCode2Name(types.PipCode(k))
				if name == types.Undefined {
					name = k
				}*/
		name, ok := kM[k]
		if !ok {
			name = types.Undefined
		}
		ret = append(ret, template.ItemInRatioData{
			Name:  name,
			Value: fmt.Sprintf("%.2f", v),
		})
	}
	return ret
}

func (s *CostTemplate) providerTypeRatioData(date tools.BillingDate) []template.ItemInRatioData {
	return []template.ItemInRatioData{
		{
			Name:  s.provider.StringCN(),
			Value: s.sumBillingDateAmount(date),
		},
	}
}

func (s *CostTemplate) chargeTypeRatioData(date tools.BillingDate) []template.ItemInRatioData {
	ret := make([]template.ItemInRatioData, 0)
	totalMap := make(map[cloud.SubscriptionType]float64) // key :prePaid or postPaid, val : bill
	for _, d := range date.Days {
		if val, ok := s.DaysBilling.Load(d); ok {
			for key, value := range val.(data.DailyBilling).ProductsBilling {
				if key == types.ECS.String() || key == "p_cvm" || key == "EC2 - Other" {
					for _, item := range value.Items {
						totalMap[item.SubscriptionType] += item.PretaxAmount
					}
				}
			}
			/*			for _, item := range val.(data.DailyBilling).ProductsBilling[types.ECS.String()].Items {
						totalMap[item.SubscriptionType] += item.PretaxAmount
					}*/
		}
	}
	for _, m := range date.Months {
		if val, ok := s.MonthsBilling.Load(m); ok {
			for key, value := range val.(data.MonthlyBilling).ProductsBilling {
				if key == types.ECS.String() || key == "p_cvm" || key == "EC2 - Other" {
					for _, item := range value.Items {
						totalMap[item.SubscriptionType] += item.PretaxAmount
					}
				}
			}
		}
	}
	for k, v := range totalMap {
		ret = append(ret, template.ItemInRatioData{
			Name:  k.StringCN(),
			Value: fmt.Sprintf("%.2f", v),
		})
	}
	return ret
}

func (s *CostTemplate) sumRatiosChartMidValue(items []template.ItemInRatioData) string {
	if len(items) == 0 {
		return "-"
	}
	var sum float64
	for _, i := range items {
		sum = tools.Float64Add(sum, cast.ToFloat64(i.Value))
	}
	return cast.ToString(sum)
}

func (s *CostTemplate) getLast14Days() []string {
	return s.bp.GetRecentXDaysBillingDate(14).Days
}
func (s *CostTemplate) getLast12Months() []string {
	ret := s.bp.GetRecentXMonthsBillingDate(12)
	months := ret.Months
	if len(ret.Days) != 0 {
		months = append(months, ret.Days[0][:7]) // 2022-10-10 -> 2022-10
	}
	return months
}
func (s *CostTemplate) getDayItemInSeries() []template.ItemInSeries {
	r := []template.ItemInSeries{
		template.ItemInSeries{
			Name:       "成本(本期)",
			Type:       "bar",
			YAxisIndex: 0,
			Data:       s.amountInLastXDays(15, true), // 近 14 天的每日成本
		},
		template.ItemInSeries{
			Name:       "成本(上一年同期)",
			Type:       "bar",
			YAxisIndex: 0,
			Data:       s.amountInLastXDays(15, false), // 去年同期 14 天的每日成本
		},
		template.ItemInSeries{
			Name:       "环比上一天",
			Type:       "line",
			YAxisIndex: 1,
			Data:       nil,
		},
		template.ItemInSeries{
			Name:       "同比上一年",
			Type:       "line",
			YAxisIndex: 1,
			Data:       nil,
		},
	}
	chainRatios := make([]string, 0, 14)
	for i := 0; i < len(r[0].Data)-1; i++ {
		chainRatios = append(chainRatios, tools.RatioString(r[0].Data[i], r[0].Data[i+1]))
	}
	yrOnyrRatios := make([]string, 0, 14)
	for i := 1; i < len(r[1].Data); i++ {
		yrOnyrRatios = append(yrOnyrRatios, tools.RatioString(r[1].Data[i], r[0].Data[i]))
	}
	r[2].Data = chainRatios  // 环比上一天 ["22.34", ... "45.67"]
	r[3].Data = yrOnyrRatios // 同比上一年 ["22.34", "--",... "45.67"]

	r[0].Data = r[0].Data[1:] // 成本(本期),只需保留 14 个值
	r[1].Data = r[1].Data[1:] // 成本(上一年同期),只需保留 14 个值
	// r[3].Data = r[3].Data[1:] //同比上一年,只需保留 14 个值

	return r
}
func (s *CostTemplate) getMonthItemInSeries() []template.ItemInSeries {
	r := []template.ItemInSeries{
		template.ItemInSeries{
			Name:       "成本(本期)",
			Type:       "bar",
			YAxisIndex: 0,
			Data:       s.amountInLastXMonths(13, true), // 近 12 月的每月成本
		},
		template.ItemInSeries{
			Name:       "成本(上一年同期)",
			Type:       "bar",
			YAxisIndex: 0,
			Data:       s.amountInLastXMonths(13, false), // 去年同期 12 月的每月成本
		},
		template.ItemInSeries{
			Name:       "环比上一月",
			Type:       "line",
			YAxisIndex: 1,
			Data:       nil,
		},
		template.ItemInSeries{
			Name:       "同比上一年",
			Type:       "line",
			YAxisIndex: 1,
			Data:       nil,
		},
	}
	chainRatios := make([]string, 0, 12)
	for i := 0; i < len(r[0].Data)-1; i++ {
		chainRatios = append(chainRatios, tools.RatioString(r[0].Data[i], r[0].Data[i+1]))
	}
	yrOnyrRatios := make([]string, 0, 12)
	for i := 1; i < len(r[0].Data); i++ {
		if i > len(r[1].Data) {
			yrOnyrRatios = append(yrOnyrRatios, "--")
			log.Println("E! lose data on year on year ratio")
			continue
		}
		yrOnyrRatios = append(yrOnyrRatios, tools.RatioString(r[1].Data[i-1], r[0].Data[i])) // (current - previous)/previous
	}
	r[2].Data = chainRatios  // 环比上一月 ["22.34", ... "45.67"]
	r[3].Data = yrOnyrRatios // 同比上一年 ["22.34", "--",... "45.67"]

	r[0].Data = r[0].Data[1:] // 成本(本期),只需保留 14 个值
	// r[1].Data = r[1].Data[1:] //成本(上一年同期),只需保留 14 个值
	// r[3].Data = r[3].Data[1:] //同比上一年,只需保留 14 个值

	return r
}
func (s *CostTemplate) amountInLastXDays(n int32, isRecentYear bool) []string {
	billingDate := s.bp.GetRecentXDaysBillingDate(n)
	if !isRecentYear {
		days := billingDate.Days
		previousYearDays := s.bp.GetTargetYearData(days, -1)
		billingDate.Days = previousYearDays
	}
	ret := make([]string, 0, len(billingDate.Days))
	for _, d := range billingDate.Days {
		if bill, ok := s.DaysBilling.Load(d); ok {
			ret = append(ret, fmt.Sprintf("%.2f", bill.(data.DailyBilling).TotalAmount))
		}
	}

	return ret
}
func (s *CostTemplate) amountInLastXMonths(n int32, isRecentYear bool) []string {
	billingDate := s.bp.GetRecentXMonthsBillingDate(n)
	if !isRecentYear {
		months := billingDate.Months
		days := billingDate.Days
		previousYearMonths := s.bp.GetTargetYearData(months, -1)
		previousYearDays := s.bp.GetTargetYearData(days, -1)
		billingDate.Months = previousYearMonths
		billingDate.Days = previousYearDays
	}
	ret := make([]string, 0, len(billingDate.Months))
	for _, m := range billingDate.Months {
		if bill, ok := s.MonthsBilling.Load(m); ok {
			ret = append(ret, fmt.Sprintf("%.2f", bill.(data.MonthlyBilling).TotalAmount))
		}
	}
	ret = append(ret, s.sumBillingDateAmount(tools.BillingDate{Days: billingDate.Days}))

	return ret
}
func (s *CostTemplate) getStatistics() []template.ItemInStatistics {
	yesterday := s.bp.GetRecentDayBillingDate()
	yesterdayT, _ := time.ParseInLocation("2006-01-02", yesterday.Days[0], time.Local)
	beforeYesterday := s.bp.GetPreviousDayBillingDate()
	recentMonth := s.bp.GetRecentMonthBillingDate(true)
	previousMonth := s.bp.ConvBillingDate2PreviousMonth(recentMonth)
	recentQuarter := s.bp.GetRecentQuarterBillingDate(true)
	previousQuarter := s.bp.ConvBillingDate2PreviousQuarter(recentQuarter)
	recentYear := s.bp.GetRecentYearBillingDate()
	previousYear := s.bp.GetPreviousYearBillingDate()
	statistics := []template.ItemInStatistics{
		template.ItemInStatistics{
			SCycle:     yesterdayT.Format("2006年01月02日累计"),
			SAmount:    s.sumBillingDateAmount(yesterday),
			SPreCycle:  "前一天同期",
			SPreAmount: s.sumBillingDateAmount(beforeYesterday),
			SRatio:     "",
		},
		template.ItemInStatistics{
			SCycle:     yesterdayT.Format("2006年01月累计"),
			SAmount:    s.sumBillingDateAmount(recentMonth),
			SPreCycle:  "上月同期",
			SPreAmount: s.sumBillingDateAmount(previousMonth),
			SRatio:     "",
		},
		template.ItemInStatistics{
			SCycle:     fmt.Sprintf("%d年第%d季度累计", s.bp.GetRecentYear(), s.bp.GetRecentQuarter()),
			SAmount:    s.sumBillingDateAmount(recentQuarter),
			SPreCycle:  "上季度同期",
			SPreAmount: s.sumBillingDateAmount(previousQuarter),
			SRatio:     "",
		},
		template.ItemInStatistics{
			SCycle:     fmt.Sprintf("%d年累计", s.bp.GetRecentYear()),
			SAmount:    s.sumBillingDateAmount(recentYear),
			SPreCycle:  "上年同期",
			SPreAmount: s.sumBillingDateAmount(previousYear),
			SRatio:     "",
		},
	}
	// 单独计算 s_ratio
	for i, v := range statistics {
		statistics[i].SRatio = tools.RatioString(v.SPreAmount, v.SAmount)
	}

	return statistics
}

func (s *CostTemplate) ExportCostAnalysis(ctx context.Context) error {
	dayAnalysis, err := s.FormatDayStatistics(ctx)
	if err != nil {
		return err
	}
	monthAnalysis, err := s.FormatMonthStatistics(ctx)
	if err != nil {
		return err
	}
	ad := template.AnalysisData{
		CostAnalysisByDay:   dayAnalysis,
		CostAnalysisByMonth: monthAnalysis,
	}
	c, err := template.ParseCostTemplate(ad)
	if err != nil {
		return err
	}
	// log.Printf("I! template content:%s", c)
	err = ioutil.WriteFile(constants.GetJsDataPath(), []byte(c), 0644)
	if err != nil {
		return err
	}

	log.Printf("I! ExportCostAnalysis done")
	return nil
}
func (s *CostTemplate) extractCurrencyUnit() (result string) {
	s.DaysBilling.Range(func(key, value interface{}) bool {
		for _, v := range value.(data.DailyBilling).ProductsBilling {
			if len(v.Items) > 0 {
				result = tools.CurrencyUnit(v.Items[0].Currency)
				return false
			}
		}
		return true
	})
	s.MonthsBilling.Range(func(key, value interface{}) bool {
		for _, v := range value.(data.MonthlyBilling).ProductsBilling {
			if len(v.Items) > 0 {
				result = tools.CurrencyUnit(v.Items[0].Currency)
				return false
			}
		}
		return true
	})
	return

}
