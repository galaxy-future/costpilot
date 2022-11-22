package databean

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/data"
	"github.com/galaxy-future/costpilot/internal/providers"
	"github.com/galaxy-future/costpilot/internal/services/datareader"
	"github.com/galaxy-future/costpilot/internal/types"
	"github.com/galaxy-future/costpilot/tools"
	"github.com/pkg/errors"
)

type UtilizationDataBean struct {
	cloudAccount types.CloudAccount
	provider     providers.Provider
	dataReader   *datareader.UtilizationDataReader

	dateRange     tools.BillingDate
	regionMap     map[string]string                // k->v: regionId->regionName
	regionZoneMap map[string][]string              // k->v: regionId->[]zoneId
	instanceMap   map[string][]data.InstanceDetail // k->v: regionId->[]data.InstanceDetail

	dailyCpu    sync.Map // 2022-01-02 -> data.DailyCpuUtilization
	dailyMemory sync.Map // 2022-01-02 -> data.DailyCpuUtilization

	instancesMap sync.Map // providerType+instanceId -> data.InstanceDetail

	bp *tools.BillingDatePilot

	pipeLineFunc []func(context.Context) error
}

func NewUtilization(a types.CloudAccount, t time.Time) *UtilizationDataBean {
	s := &UtilizationDataBean{
		cloudAccount: a,
		dateRange:    tools.BillingDate{},
		bp:           tools.NewBillDatePilot().SetNowT(t),
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

// newRegionProvider create provider by new region
func (s *UtilizationDataBean) newRegionProvider(regionId string) providers.Provider {
	var err error
	p, err := providers.GetProvider(s.cloudAccount.Provider, s.cloudAccount.AK, s.cloudAccount.SK, regionId)
	if err != nil {
		log.Printf("E! newRegionProvider failed: %v\n", err)
	}
	return p
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

func (s *UtilizationDataBean) loadAllZones(ctx context.Context) error {
	var err error
	regionMap := make(map[string]string)
	regionZonesMap := make(map[string][]string)
	regionMap, err = s.dataReader.GetAllRegionMap(ctx)
	if err != nil {
		log.Printf("E! loadAllZones.GetAllRegionMap:%v", err)
		return err
	}
	if len(regionMap) == 0 {
		return nil
	}
	for region, _ := range regionMap {
		if !strings.Contains(region, "beijing") && !strings.Contains(region, "shanghai") && !strings.Contains(region, "guangzhou") {
			continue
		}
		p := s.newRegionProvider(region)
		zoneMap, err := s.dataReader.GetAllZoneMap(ctx, p, region, true)
		if err != nil {
			log.Printf("E! loadAllZones.GetAllZoneMap:%v", err)
			return err
		}
		for id, _ := range zoneMap {
			regionZonesMap[region] = append(regionZonesMap[region], id)
		}
	}

	s.regionZoneMap = regionZonesMap
	log.Printf("I! loadAllZones success,len=%d", len(regionZonesMap))
	return nil
}

func (s *UtilizationDataBean) getAllInstances(ctx context.Context) error {
	for regionId, zoneIdList := range s.regionZoneMap {
		p := s.newRegionProvider(regionId)
		instanceList, err := s.dataReader.GetInstanceByZones(ctx, p, zoneIdList)
		if err != nil {
			log.Printf("E! getAllInstances.GetInstanceByZones:%v", err)
			return err
		}
		s.instanceMap[regionId] = append(s.instanceMap[regionId], instanceList...)
	}
	log.Printf("I! getAllInstances success,len=%d", len(s.instanceMap))
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

	cpuData, err := dataReader.GetDaysCpuUtilization(ctx, nil, []string{}, days...)
	if err != nil {
		return err
	}
	for _, v := range cpuData {
		s.dailyCpu.LoadOrStore(v.Day, v)
	}

	return nil
}

func (s *UtilizationDataBean) fetchCpuUtilizationByInstanceIds(ctx context.Context) error {
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
	log.Printf("I! fetchCpuUtilizationByInstanceIds days: %v", days)

	for regionId, instanceList := range s.instanceMap {
		var ids = make([]string, len(instanceList))
		for _, i := range instanceList {
			ids = append(ids, i.InstanceId)
		}
		p := s.newRegionProvider(regionId)
		cpuData, err := dataReader.GetDaysCpuUtilization(ctx, p, ids, days...)
		if err != nil {
			return err
		}
		for _, v := range cpuData {
			d, ok := s.dailyCpu.Load(v.Day)
			if ok {
				cpuDay := d.(data.DailyCpuUtilization)
				cpuDay.Utilization = append(cpuDay.Utilization, v.Utilization...)
			} else {
				s.dailyCpu.LoadOrStore(v.Day, v)
			}
		}
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
	memoryData, err := s.dataReader.GetDaysMemoryUtilization(ctx, nil, []string{}, days...)
	if err != nil {
		return err
	}

	for _, v := range memoryData {
		s.dailyMemory.LoadOrStore(v.Day, v)
	}
	return nil
}

func (s *UtilizationDataBean) fetchMemoryUtilizationByInstanceIds(ctx context.Context) error {
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
	log.Printf("I! fetchMemoryUtilizationByInstanceIds days: %v", days)

	for regionId, instanceList := range s.instanceMap {
		var ids = make([]string, len(instanceList))
		for _, i := range instanceList {
			ids = append(ids, i.InstanceId)
		}
		p := s.newRegionProvider(regionId)
		memoryData, err := s.dataReader.GetDaysMemoryUtilization(ctx, p, ids, days...)
		if err != nil {
			return err
		}
		for _, v := range memoryData {
			d, ok := s.dailyCpu.Load(v.Day)
			if ok {
				memoryDay := d.(data.DailyMemoryUtilization)
				memoryDay.Utilization = append(memoryDay.Utilization, v.Utilization...)
				s.dailyMemory.LoadOrStore(v.Day, memoryDay)
			} else {
				s.dailyMemory.LoadOrStore(v.Day, v)
			}
		}
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

func (s *UtilizationDataBean) getRecentInstanceListFromLocal(ctx context.Context) error {
	return nil
}

func (s *UtilizationDataBean) GetUtilizationAnalysisPipeLine() []func(context.Context) error {
	var pipeLine []func(context.Context) error

	// 有的厂商要先去拉取所有实例，然后才能去抓监控数据
	if s.provider.ProviderType() != cloud.AlibabaCloud {
		pipeLine = append(pipeLine, s.loadAllZones, s.getAllInstances)
	}

	// 通用处理
	pipeLine = append(pipeLine,
		s.loadRegionMap,
		s.getRecentDay,
		s.getPreviousDay,
		s.getRecent14DaysDate,
	)

	if s.provider.ProviderType() == cloud.AlibabaCloud {
		pipeLine = append(pipeLine, s.fetchCpuUtilization, s.fetchMemoryUtilization, s.fetchRecentInstanceList)
	} else {
		pipeLine = append(pipeLine,
			s.fetchCpuUtilizationByInstanceIds,
			s.fetchMemoryUtilizationByInstanceIds,
			s.getRecentInstanceListFromLocal,
		)
	}

	return pipeLine
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
