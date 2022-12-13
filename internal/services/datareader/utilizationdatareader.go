package datareader

import (
	"context"
	"log"
	"time"

	"github.com/galaxy-future/costpilot/internal/data"
	"github.com/galaxy-future/costpilot/internal/providers"
	"github.com/galaxy-future/costpilot/internal/providers/types"
	"github.com/galaxy-future/costpilot/tools"
	"golang.org/x/sync/errgroup"
)

type UtilizationDataReader struct {
	_provider providers.Provider
}

func NewUtilization(p providers.Provider) *UtilizationDataReader {
	return &UtilizationDataReader{
		_provider: p,
	}
}

// GetDailyCpuUtilization
func (s *UtilizationDataReader) GetDailyCpuUtilization(ctx context.Context, day string, p providers.Provider, instanceIds []string) (data.DailyCpuUtilization, error) {
	if !tools.IsValidDayDate(day) {
		log.Printf("W! invalid day[%v]\n", day)
		return data.DailyCpuUtilization{}, nil
	}

	startTime, err := time.ParseInLocation("2006-01-02", day, time.Local)
	if err != nil {
		return data.DailyCpuUtilization{}, nil
	}
	endTime := startTime.AddDate(0, 0, +1)

	provider := s._provider
	if p != nil {
		provider = p
	}
	resp, err := provider.DescribeMetricList(ctx, types.DescribeMetricListRequest{
		MetricName: types.MetricItemCPUUtilization,
		Period:     "86400", // 一天
		StartTime:  startTime,
		EndTime:    endTime,
		Filter:     types.MetricListInstanceFilter{InstanceIds: instanceIds},
	})
	if err != nil {
		return data.DailyCpuUtilization{}, err
	}

	result := data.DailyCpuUtilization{
		Provider: provider.ProviderType(),
		Day:      day,
	}

	for _, v := range resp.List {
		result.Utilization = append(result.Utilization, data.InstanceCpuUtilization{
			InstanceId:      v.InstanceId,
			UsedUtilization: v.Average,
		})
	}
	return result, nil
}

func (s *UtilizationDataReader) GetDaysCpuUtilization(ctx context.Context, p providers.Provider, instanceIds []string, days ...string) ([]data.DailyCpuUtilization, error) {
	var result []data.DailyCpuUtilization
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
				log.Printf("I! Canceled GetDays[%s]\n", d)
				return nil
			default:
				res, err := s.GetDailyCpuUtilization(ctx, d, p, instanceIds)
				if err != nil {
					return err
				}
				log.Printf("I! GetDailyCpuUtilization [%v]", d)
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
	return result, nil
}

func (s *UtilizationDataReader) GetDailyMemoryUtilization(ctx context.Context, p providers.Provider, instanceIds []string, day string) (data.DailyMemoryUtilization, error) {
	if !tools.IsValidDayDate(day) {
		log.Printf("W! invalid day[%v]\n", day)
		return data.DailyMemoryUtilization{}, nil
	}

	startTime, err := time.ParseInLocation("2006-01-02", day, time.Local)
	if err != nil {
		return data.DailyMemoryUtilization{}, nil
	}
	endTime := startTime.AddDate(0, 0, +1)

	provider := s._provider
	if p != nil {
		provider = p
	}
	resp, err := provider.DescribeMetricList(ctx, types.DescribeMetricListRequest{
		MetricName: types.MetricItemMemoryUsedUtilization,
		Period:     "86400", // 一天
		StartTime:  startTime,
		EndTime:    endTime,
		Filter: types.MetricListInstanceFilter{
			InstanceIds: instanceIds,
		},
	})
	if err != nil {
		return data.DailyMemoryUtilization{}, err
	}

	result := data.DailyMemoryUtilization{
		Provider: provider.ProviderType(),
		Day:      day,
	}

	for _, v := range resp.List {
		result.Utilization = append(result.Utilization, data.InstanceMemoryUtilization{
			InstanceId:      v.InstanceId,
			UsedUtilization: v.Average,
		})
	}
	return result, nil
}

func (s *UtilizationDataReader) GetDaysMemoryUtilization(ctx context.Context, p providers.Provider, instanceIds []string, days ...string) ([]data.DailyMemoryUtilization, error) {
	var result []data.DailyMemoryUtilization
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
				log.Printf("I! Canceled GetDays[%s]\n", d)
				return nil
			default:
				res, err := s.GetDailyMemoryUtilization(ctx, p, instanceIds, d)
				if err != nil {
					log.Printf("E! GetDailyCpuUtilization [%v], error=[%v]", d, err)
					return err
				}
				log.Printf("I! GetDailyMemoryUtilization [%s]", d)
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
	return result, nil
}

func (s *UtilizationDataReader) GetInstanceList(ctx context.Context, instanceIdList ...string) ([]data.InstanceDetail, error) {
	resp, err := s._provider.QueryAvailableInstances(ctx, types.QueryAvailableInstancesRequest{
		InstanceIdList: instanceIdList,
	})
	if err != nil {
		log.Printf("W! %s.GetInstanceList.QueryAvailableInstances:%v", s._provider.ProviderType().String(), err)
		return []data.InstanceDetail{}, err
	}

	availableIdMap := make(map[string]bool)
	result := make([]data.InstanceDetail, 0, len(instanceIdList))
	for _, instance := range resp.List {
		i := data.InstanceDetail{
			Provider:         s._provider.ProviderType(),
			InstanceId:       instance.InstanceId,
			RegionId:         instance.RegionId,
			SubscriptionType: instance.SubscriptionType,
		}
		availableIdMap[instance.InstanceId] = true
		result = append(result, i)
	}
	var invalidIdList []string
	for _, id := range instanceIdList {
		if !availableIdMap[id] {
			invalidIdList = append(invalidIdList, id)
		}
	}
	dateTool := tools.NewBillDatePilot().SetNowT(time.Now())
	cycle := dateTool.GetRecentMonth()
	for _, i := range invalidIdList {
		resp2, err2 := s._provider.DescribeInstanceBill(ctx, types.DescribeInstanceBillRequest{
			BillingCycle: cycle,
			Granularity:  types.Monthly,
			InstanceId:   i,
		}, false)
		if err2 != nil {
			log.Printf("W! %s.GetInstanceList.DescribeInstanceBill:%v", s._provider.ProviderType().String(), err2)
			return []data.InstanceDetail{}, err2
		}
		if len(resp2.Items) == 0 {
			log.Printf("W! %s.GetInstanceList: can not find %s instance detail", s._provider.ProviderType().String(), i)
			continue
		}
		detail := resp2.Items[0]
		result = append(result, data.InstanceDetail{
			Provider:         s._provider.ProviderType(),
			InstanceId:       i,
			RegionName:       detail.Region,
			SubscriptionType: detail.SubscriptionType,
		})
	}
	return result, nil
}

// GetAllRegionMap  k->v: regionId->regionName
func (s *UtilizationDataReader) GetAllRegionMap(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string)
	resp, err := s._provider.DescribeRegions(ctx, types.DescribeRegionsRequest{
		ResourceType: types.ResourceTypeInstance,
		Language:     types.RegionLanguageZHCN,
	})
	if err != nil {
		return result, err
	}
	for _, region := range resp.List {
		result[region.RegionId] = region.LocalName
	}
	return result, nil
}

func (s *UtilizationDataReader) GetInstanceByRegionProvider(ctx context.Context, p providers.Provider, regionId string) ([]data.InstanceDetail, error) {
	var result []data.InstanceDetail

	resp, err := p.DescribeInstances(ctx, types.DescribeInstancesRequest{})
	if err != nil {
		return result, err
	}
	for _, i := range resp.List {
		result = append(result, data.InstanceDetail{
			Provider:         p.ProviderType(),
			InstanceId:       i.InstanceId,
			RegionId:         regionId,
			SubscriptionType: i.SubscriptionType,
		})
	}

	return result, nil
}
