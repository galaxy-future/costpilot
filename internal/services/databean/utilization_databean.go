package databean

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/galaxy-future/costpilot/internal/data"
	"github.com/galaxy-future/costpilot/internal/providers"
	"github.com/galaxy-future/costpilot/internal/services/datareader"
	"github.com/galaxy-future/costpilot/internal/types"
	"github.com/galaxy-future/costpilot/tools"
	"github.com/pkg/errors"
)

type UtilizationDataBean struct {
	provider   providers.Provider
	dataReader *datareader.UtilizationDataReader

	dateRange tools.BillingDate
	regionMap map[string]string

	dailyCpu    sync.Map // 2022-01-02 -> data.DailyCpuUtilization
	dailyMemory sync.Map // 2022-01-02 -> data.DailyCpuUtilization

	instancesMap sync.Map // providerType+instanceId -> data.InstanceDetail

	bp *tools.BillingDatePilot

	pipeLineFunc []func(context.Context) error
}

func NewUtilization(a types.CloudAccount, t time.Time) *UtilizationDataBean {
	s := &UtilizationDataBean{
		dateRange: tools.BillingDate{},
		bp:        tools.NewBillDatePilot().SetNowT(t),
	}
	s.initProvider(a)
	s.initDataReader()
	return s
}

// initProvider
func (s *UtilizationDataBean) initProvider(a types.CloudAccount) *UtilizationDataBean {
	var err error
	s.provider, err = providers.GetProvider(a.Provider, a.AK, a.SK, a.RegionID)
	if err != nil {
		log.Printf("E! init provider failed: %v\n", err)
	}
	return s
}

func (s *UtilizationDataBean) initDataReader() {
	s.dataReader = datareader.NewUtilization(s.provider)
}

func (s *UtilizationDataBean) loadRegionMap(ctx context.Context) error {
	regionMap, err := s.dataReader.GetAllRegionMap(ctx)
	if err != nil {
		log.Printf("E! loadRegionMap:%v", err)
		return err
	}
	s.regionMap = regionMap
	log.Printf("I! loadRegionMap success,len=%d", len(regionMap))
	return nil
}

func (s *UtilizationDataBean) getRecentDay(ctx context.Context) error {
	dateList := s.bp.GetRecentDayBillingDate()
	if err := s.AddDate(ctx, dateList); err != nil {
		return err
	}
	return nil
}

func (s *UtilizationDataBean) getPreviousDay(ctx context.Context) error {
	dateList := s.bp.GetPreviousDayBillingDate()
	if err := s.AddDate(ctx, dateList); err != nil {
		return err
	}
	return nil
}

// getRecent14DaysDate today is not included
func (s *UtilizationDataBean) getRecent14DaysDate(ctx context.Context) error {
	dateList := s.bp.GetRecentXDaysBillingDate(14)
	if err := s.AddDate(ctx, dateList); err != nil {
		return err
	}
	return nil
}

func (s *UtilizationDataBean) AddDate(_ context.Context, dateList tools.BillingDate) error {
	s.dateRange.Days = tools.Union(s.dateRange.Days, dateList.Days)
	return nil
}

func (s *UtilizationDataBean) fetchCpuUtilization(ctx context.Context) error {
	b := s.dateRange
	dataReader := datareader.NewUtilization(s.provider)
	var days []string

	for _, v := range b.Days {
		if _, ok := s.dailyCpu.Load(v); !ok {
			days = append(days, v)
		}
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i] < days[j]
	})
	log.Printf("I! fetchCpuUtilization days: %v", days)

	cpuData, err := dataReader.GetDaysCpuUtilization(ctx, days...)
	if err != nil {
		return err
	}
	for _, v := range cpuData {
		s.dailyCpu.LoadOrStore(v.Day, v)
	}

	return nil
}

func (s *UtilizationDataBean) fetchMemoryUtilization(ctx context.Context) error {
	b := s.dateRange
	var days []string

	for _, v := range b.Days {
		if _, ok := s.dailyMemory.Load(v); !ok {
			days = append(days, v)
		}
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i] < days[j]
	})
	log.Printf("I! fetchMemoryUtilization days: %v", days)
	memoryData, err := s.dataReader.GetDaysMemoryUtilization(ctx, days...)
	if err != nil {
		return err
	}

	for _, v := range memoryData {
		s.dailyMemory.LoadOrStore(v.Day, v)
	}
	return nil
}

func (s *UtilizationDataBean) fetchRecentInstanceList(ctx context.Context) error {
	d := s.bp.GetRecentDayBillingDate()
	recentDay := d.Days[0]
	v, ok := s.dailyCpu.Load(recentDay)
	if !ok {
		return errors.New("no instance running")
	}
	if len(s.regionMap) == 0 {
		return errors.New("you must reload region map firstly")
	}
	vv := v.(data.DailyCpuUtilization)
	var idList []string
	for _, u := range vv.Utilization {
		idList = append(idList, u.InstanceId)
	}
	instanceList, err := s.dataReader.GetInstanceList(ctx, idList...)
	if err != nil {
		log.Printf("E! GetInstanceList:%v", err)
		return err
	}
	log.Printf("I! fetchRecentInstanceList len=%d", len(instanceList))
	for _, detail := range instanceList {
		k := fmt.Sprintf("%s:%s", s.provider.ProviderType(), detail.InstanceId)
		if len(detail.RegionName) > 0 {
			s.instancesMap.Store(k, detail)
			continue
		}
		var regionName string
		regionName, ok = s.regionMap[detail.RegionId]
		if !ok {
			regionName = "未知"
		}
		detail.RegionName = regionName
		s.instancesMap.Store(k, detail)
	}
	return nil
}

func (s *UtilizationDataBean) GetUtilizationAnalysisPipeLine() []func(context.Context) error {
	return []func(context.Context) error{
		s.loadRegionMap,

		s.getRecentDay,
		s.getPreviousDay,
		s.getRecent14DaysDate,

		s.fetchCpuUtilization,
		s.fetchMemoryUtilization,

		s.fetchRecentInstanceList,
	}
}

func (s *UtilizationDataBean) RunPipeline(ctx context.Context) error {
	var err error
	for _, f := range s.GetUtilizationAnalysisPipeLine() {
		err = f(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *UtilizationDataBean) GetUtilizationMap() (*sync.Map, *sync.Map, *sync.Map) {
	return &s.dailyCpu, &s.dailyMemory, &s.instancesMap
}
