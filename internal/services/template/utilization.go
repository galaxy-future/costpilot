package template

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/galaxy-future/costpilot/internal/constants"
	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/data"
	"github.com/galaxy-future/costpilot/internal/template"
	"github.com/galaxy-future/costpilot/tools"
	"github.com/spf13/cast"
)

const _invalidValue = "--"

type UtilizationTemplate struct {
	bp *tools.BillingDatePilot

	CpuUtilization     *sync.Map // key : day , val : []data.DailyCpuUtilization
	MemoryUtilization  *sync.Map // key : day , val : []data.DailyMemoryUtilization
	RecentInstanceList []data.InstanceDetail
}

func NewUtilization(t time.Time) *UtilizationTemplate {
	return &UtilizationTemplate{
		bp: tools.NewBillDatePilot().SetNowT(t),
	}
}

func (s *UtilizationTemplate) AssignData(dailyCpuProviders, dailyMemoryProviders, recentInstancesProviders []*sync.Map) {
	// 以账号为维度，遍历各账号下每天资源使用率
	for _, dailyCpuProvider := range dailyCpuProviders {
		if s.CpuUtilization == nil {
			var t sync.Map
			dailyCpuProvider.Range(func(key, value interface{}) bool {
				v := value.(data.DailyCpuUtilization)
				t.Store(key, []data.DailyCpuUtilization{v})
				return true
			})
			s.CpuUtilization = &t
			continue
		}
		dailyCpuProvider.Range(func(key, value interface{}) bool {
			v := value.(data.DailyCpuUtilization)
			val, ok := s.CpuUtilization.Load(key)
			var cpuUtilization []data.DailyCpuUtilization
			if ok {
				cpuUtilization = val.([]data.DailyCpuUtilization)
			}
			cpuUtilization = append(cpuUtilization, data.DailyCpuUtilization{
				Provider:    v.Provider,
				Day:         v.Day,
				Utilization: v.Utilization,
			})
			s.CpuUtilization.Store(key, cpuUtilization)
			return true
		})
	}

	for _, dailyMemoryProvider := range dailyMemoryProviders {
		if s.MemoryUtilization == nil {
			var t sync.Map
			dailyMemoryProvider.Range(func(key, value interface{}) bool {
				v := value.(data.DailyMemoryUtilization)
				t.Store(key, []data.DailyMemoryUtilization{v})
				return true
			})
			s.MemoryUtilization = &t
			continue
		}
		dailyMemoryProvider.Range(func(key, value interface{}) bool {
			v := value.(data.DailyMemoryUtilization)
			val, ok := s.MemoryUtilization.Load(key)
			var memoryUtilization []data.DailyMemoryUtilization
			if ok {
				memoryUtilization = val.([]data.DailyMemoryUtilization)
			}
			memoryUtilization = append(memoryUtilization, data.DailyMemoryUtilization{
				Provider:    v.Provider,
				Day:         v.Day,
				Utilization: v.Utilization,
			})
			s.MemoryUtilization.Store(key, memoryUtilization)
			return true
		})
	}

	for _, provider := range recentInstancesProviders {
		provider.Range(func(key, value interface{}) bool {
			val := value.(data.InstanceDetail)
			s.RecentInstanceList = append(s.RecentInstanceList, data.InstanceDetail{
				Provider:         val.Provider,
				InstanceId:       val.InstanceId,
				RegionId:         val.RegionId,
				RegionName:       val.RegionName,
				SubscriptionType: val.SubscriptionType,
			})
			return true
		})
	}
}

func (s *UtilizationTemplate) averagingCpuUsedRatio(date tools.BillingDate) string {
	var (
		total float64
		num   = 0
	)
	for _, d := range date.Days {
		if val, ok := s.CpuUtilization.Load(d); ok {
			for _, list := range val.([]data.DailyCpuUtilization) {
				for _, i := range list.Utilization {
					num++
					total += i.UsedUtilization
				}
			}
		}
	}
	if num == 0 {
		return _invalidValue
	}
	return fmt.Sprintf("%.2f", total/float64(num))
}

// averagingCpuUsedRatio
func (s *UtilizationTemplate) averagingMemoryUsedRatio(date tools.BillingDate) string {
	var (
		total float64
		num   = 0
	)
	for _, d := range date.Days {
		if val, ok := s.MemoryUtilization.Load(d); ok {
			for _, list := range val.([]data.DailyMemoryUtilization) {
				for _, i := range list.Utilization {
					num++
					total += i.UsedUtilization
				}
			}
		}
	}
	if num == 0 {
		return _invalidValue
	}

	return fmt.Sprintf("%.2f", total/float64(num))
}

func (s *UtilizationTemplate) sumSvrNum(date tools.BillingDate) string {
	var sum int
	for _, d := range date.Days {
		if val, ok := s.CpuUtilization.Load(d); ok {
			for _, list := range val.([]data.DailyCpuUtilization) {
				sum += len(list.Utilization)
			}
		}
	}

	return fmt.Sprintf("%d", sum)
}

func (s *UtilizationTemplate) getStatistics() []template.UtilizeAnalysisStatisticsItem {
	yesterday := s.bp.GetRecentDayBillingDate()
	beforeYesterday := s.bp.GetPreviousDayBillingDate()

	return []template.UtilizeAnalysisStatisticsItem{
		{
			SCycle:     "CPU 平均利用率",
			SAmount:    s.averagingCpuUsedRatio(yesterday),
			SUnit:      "%",
			SPreCycle:  "前日数据",
			SPreAmount: s.averagingCpuUsedRatio(beforeYesterday),
			SPreUnit:   "%",
		},
		{
			SCycle:     "内存平均利用率",
			SAmount:    s.averagingMemoryUsedRatio(yesterday),
			SUnit:      "%",
			SPreCycle:  "前日数据",
			SPreAmount: s.averagingMemoryUsedRatio(beforeYesterday),
			SPreUnit:   "%",
		},
		{
			SCycle:     "云服务器",
			SAmount:    s.sumSvrNum(yesterday),
			SUnit:      "台",
			SPreCycle:  "前日数据",
			SPreAmount: s.sumSvrNum(beforeYesterday),
			SPreUnit:   "台",
		},
	}
}

func (s *UtilizationTemplate) extractSubscriptionType() []template.ItemInRatioData {
	radioMap := make(map[cloud.SubscriptionType]int)
	for _, i := range s.RecentInstanceList {
		radioMap[i.SubscriptionType]++
	}
	var itemData []template.ItemInRatioData
	for subscriptionType, i := range radioMap {
		itemData = append(itemData, template.ItemInRatioData{
			Name:  subscriptionType.StringCN(),
			Value: fmt.Sprintf("%d", i),
		})
	}

	return itemData
}

func (s *UtilizationTemplate) extractRegion() []template.ItemInRatioData {
	radioMap := make(map[string]int)
	for _, i := range s.RecentInstanceList {
		k := fmt.Sprintf("%s-%s", i.Provider.StringCN(), i.RegionName)
		radioMap[k]++
	}
	var itemData []template.ItemInRatioData
	for k, v := range radioMap {
		itemData = append(itemData, template.ItemInRatioData{
			Name:  k,
			Value: fmt.Sprintf("%d", v),
		})
	}

	return itemData
}

func (s *UtilizationTemplate) extractProvider() []template.ItemInRatioData {
	radioMap := make(map[cloud.Provider]int)
	for _, i := range s.RecentInstanceList {
		radioMap[i.Provider]++
	}
	var itemData []template.ItemInRatioData
	for p, n := range radioMap {
		itemData = append(itemData, template.ItemInRatioData{
			Name:  p.StringCN(),
			Value: fmt.Sprintf("%d", n),
		})
	}

	return itemData
}

func (s *UtilizationTemplate) getLast14Days() []string {
	return s.bp.GetRecentXDaysBillingDate(14).Days
}

// getDailyDistributionByCpuRadio
func (s *UtilizationTemplate) getDailyDistributionByCpuRadio(start, end float64) []string {
	dateRange := s.bp.GetRecentXDaysBillingDate(14)
	ret := make([]string, 0, len(dateRange.Days))
	counter := make(map[string]int)
	for _, d := range dateRange.Days {
		if v, ok := s.CpuUtilization.Load(d); ok {
			list := v.([]data.DailyCpuUtilization)
			for _, val := range list {
				for _, utilization := range val.Utilization {
					if utilization.UsedUtilization >= start && utilization.UsedUtilization < end {
						counter[d]++
					}
				}
			}
		}
	}
	for _, d := range dateRange.Days {
		n, ok := counter[d]
		var valDay = _invalidValue
		if ok {
			valDay = fmt.Sprintf("%d", n)
		}
		ret = append(ret, valDay)
	}

	return ret
}

func (s *UtilizationTemplate) getCpuUsedInLastXDays(n int32) []string {
	dateRange := s.bp.GetRecentXDaysBillingDate(n)
	ret := make([]string, 0, len(dateRange.Days))
	for _, d := range dateRange.Days {
		ret = append(ret, s.averagingCpuUsedRatio(tools.BillingDate{Days: []string{d}}))
	}

	return ret
}

func (s *UtilizationTemplate) getMemoryUsedInLastXDays(n int32) []string {
	dateRange := s.bp.GetRecentXDaysBillingDate(n)
	ret := make([]string, 0, len(dateRange.Days))
	for _, d := range dateRange.Days {
		ret = append(ret, s.averagingMemoryUsedRatio(tools.BillingDate{Days: []string{d}}))
	}

	return ret
}

func (s *UtilizationTemplate) getUtilizeTrendSeries() []template.UtilizeAnalysisItemInSeries {
	return []template.UtilizeAnalysisItemInSeries{
		{
			Name: "CPU",
			Data: s.getCpuUsedInLastXDays(14),
		},
		{
			Name: "内存",
			Data: s.getMemoryUsedInLastXDays(14),
		},
	}
}

func (s *UtilizationTemplate) getCpuTrendSeries() []template.UtilizeAnalysisItemInSeries {
	return []template.UtilizeAnalysisItemInSeries{
		{
			Name: "0%~20%",
			Data: s.getDailyDistributionByCpuRadio(0, 20),
		},
		{
			Name: "20%~40%",
			Data: s.getDailyDistributionByCpuRadio(20, 40),
		},
		{
			Name: "40%~60%",
			Data: s.getDailyDistributionByCpuRadio(40, 60),
		},
		{
			Name: "60%~80%",
			Data: s.getDailyDistributionByCpuRadio(60, 80),
		},
		{
			Name: "80%以上",
			Data: s.getDailyDistributionByCpuRadio(80, 100),
		},
	}
}

func (s *UtilizationTemplate) sumRatiosChartMidValue(items []template.ItemInRatioData) string {
	if len(items) == 0 {
		return "-"
	}
	var sum float64
	for _, i := range items {
		sum = tools.Float64Add(sum, cast.ToFloat64(i.Value))
	}
	return cast.ToString(sum)
}

func (s *UtilizationTemplate) Assemble(_ context.Context) template.UtilizeAnalysis {
	recentDay := s.bp.GetRecentDayBillingDate()
	utilizeAnalysisByDay := template.UtilizeAnalysisByDay{
		ViewType:   "day",
		DataCycle:  recentDay.Days[0] + " 23:59:59",
		Statistics: s.getStatistics(),
	}

	utilizeAnalysisByDay.Ratios = []template.ItemInRatios{
		{
			Chart: template.ChartInRatios{
				ID:      "chargeTypeRatio",
				Title:   "服务器付费类型",
				MidUnit: "台",
				Data:    s.extractSubscriptionType(),
			},
		},
		{
			Chart: template.ChartInRatios{
				ID:      "providerTypeRatio",
				Title:   "服务器云厂商分布",
				MidUnit: "台",
				Data:    s.extractProvider(),
			},
		},
		{
			Chart: template.ChartInRatios{
				ID:      "regionTypeRatio",
				Title:   "云服务器地域分布",
				MidUnit: "台",
				Data:    s.extractRegion(),
			},
		},
	}

	utilizeAnalysisByDay.CpuTrend = &template.UtilizeAnalysisCpuTrend{
		Chart: template.ChartCpuTrend{
			ID:     "cpuTrend",
			Title:  "CPU 利用率分布",
			XData:  s.getLast14Days(),
			Series: s.getCpuTrendSeries(),
			YTitle: []string{"机器数 (个)"},
			TooltipUnit: template.TooltipUnit{
				Bar: "个",
			},
			Style: template.ChartTrendStyle{Stack: "total"},
		},
	}
	utilizeAnalysisByDay.UtilizeTrend = &template.UtilizeAnalysisUtilizeTrend{
		Chart: template.ChartUtilizeTrend{
			ID:     "utilizeTrend",
			Title:  "利用率走势",
			XData:  s.getLast14Days(),
			Series: s.getUtilizeTrendSeries(),
			YTitle: []string{"利用率（%）"},
			TooltipUnit: template.TooltipUnit{
				Line: "%",
			},
		},
	}
	// 单独计算 MidValue
	for i, v := range utilizeAnalysisByDay.Ratios {
		utilizeAnalysisByDay.Ratios[i].Chart.MidValue = s.sumRatiosChartMidValue(v.Chart.Data)
	}

	for i, v := range utilizeAnalysisByDay.Statistics {
		utilizeAnalysisByDay.Statistics[i].SRatio = tools.RatioString(v.SPreAmount, v.SAmount)
	}

	return template.UtilizeAnalysis{AnalysisByDay: utilizeAnalysisByDay}
}

func (s *UtilizationTemplate) Export(_ context.Context, ua template.UtilizeAnalysis) error {
	c, err := template.ParseUtilizeAnalysisTemplate(ua)
	if err != nil {
		return err
	}
	jsContent, err := os.ReadFile(constants.GetJsDataPath())
	if err != nil {
		return err
	}
	err = os.WriteFile(constants.GetJsDataPath(), append(jsContent, []byte(c)...), 0644)
	if err != nil {
		return err
	}

	log.Printf("I! UtilizeAnalysis done")
	return nil
}
