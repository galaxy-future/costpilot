package datareader

import (
	"context"
	"log"
	"time"

	"github.com/galayx-future/costpilot/internal/data"
	"github.com/galayx-future/costpilot/internal/providers"
	"github.com/galayx-future/costpilot/internal/providers/types"
	"github.com/galayx-future/costpilot/tools"
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
func (s *UtilizationDataReader) GetDailyCpuUtilization(ctx context.Context, day string) (data.DailyCpuUtilization, error) {
	if !tools.IsValidDayDate(day) {
		log.Printf("W! invalid day[%v]\n", day)
		return data.DailyCpuUtilization{}, nil
	}

	startTime, err := time.ParseInLocation("2006-01-02", day, time.Local)
	if err != nil {
		return data.DailyCpuUtilization{}, nil
	}
	endTime := startTime.AddDate(0, 0, +1)

	resp, err := s._provider.DescribeMetricList(ctx, types.DescribeMetricListRequest{
		MetricName: types.MetricItemCPUUtilization,
		Period:     "86400", // 一天
		StartTime:  startTime,
		EndTime:    endTime,
	})
	if err != nil {
		return data.DailyCpuUtilization{}, err
	}

	result := data.DailyCpuUtilization{
		Provider: s._provider.ProviderType(),
		Day:      day,
	}

	for _, v := range resp.List {
		result.Utilization = append(result.Utilization, data.InstanceCpuUtilization{
			InstanceId:      v.InstanceId,
			UsedUtilization: v.Average,
		})
	}
	return result, nil
	//
	// resp2, err2 := s._provider.DescribeMetricList(ctx, types.DescribeMetricListRequest{
	// 	MetricName: types.MetricItemMemoryUsedUtilization,
	// 	Period:     "86400", // 一天
	// 	StartTime:  startTime,
	// 	EndTime:    endTime,
	// })
	// if err2 != nil {
	// 	return data.ResourceUtilization{}, err
	// }
	// for _, v := range resp2.List {
	// 	item := sampleMap[v.Instance]
	// 	item.MemoryUtilization = v.Average
	// }
	// if len(sampleMap) == 0 {
	// 	return sampleMap, nil
	// }
	//
	// for instanceId, _ := range sampleMap {
	// 	item := sampleMap[instanceId]
	// 	resp3, err3 := s._provider.DescribeInstanceAttribute(ctx, types.DescribeInstanceAttributeRequest{InstanceId: instanceId})
	// 	if err3 != nil {
	// 		return result, err
	// 	}
	// 	item.RegionId = resp3.InstanceId
	// 	item.SubscriptionType = resp3.SubscriptionType
	// }
	//
	// regionIdMap := make(map[string]string)
	// resp4, err4 := s._provider.DescribeRegions(ctx, types.DescribeRegionsRequest{
	// 	ResourceType: types.ResourceTypeInstance,
	// 	Language:     types.RegionLanguageZHCN,
	// })
	// if err4 != nil {
	// 	return result, err
	// }
	// for _, region := range resp4.List {
	// 	regionIdMap[region.RegionId] = region.LocalName
	// }
	//
	// for _, sample := range sampleMap {
	// 	if regionName, ok := regionIdMap[sample.RegionId]; ok {
	// 		sample.RegionName = regionName
	// 	}
	// }
	// result.Utilization[] =
}

func (s *UtilizationDataReader) GetDaysCpuUtilization(ctx context.Context, days ...string) ([]data.DailyCpuUtilization, error) {
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
				res, err := s.GetDailyCpuUtilization(ctx, d)
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
	return result, nil
}

func (s *UtilizationDataReader) GetDailyMemoryUtilization(ctx context.Context, day string) (data.DailyMemoryUtilization, error) {
	if !tools.IsValidDayDate(day) {
		log.Printf("W! invalid day[%v]\n", day)
		return data.DailyMemoryUtilization{}, nil
	}

	startTime, err := time.ParseInLocation("2006-01-02", day, time.Local)
	if err != nil {
		return data.DailyMemoryUtilization{}, nil
	}
	endTime := startTime.AddDate(0, 0, +1)

	resp, err := s._provider.DescribeMetricList(ctx, types.DescribeMetricListRequest{
		MetricName: types.MetricItemMemoryUsedUtilization,
		Period:     "86400", // 一天
		StartTime:  startTime,
		EndTime:    endTime,
	})
	if err != nil {
		return data.DailyMemoryUtilization{}, err
	}

	result := data.DailyMemoryUtilization{
		Provider: s._provider.ProviderType(),
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

func (s *UtilizationDataReader) GetDaysMemoryUtilization(ctx context.Context, days ...string) ([]data.DailyMemoryUtilization, error) {
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
				res, err := s.GetDailyMemoryUtilization(ctx, d)
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
	return result, nil
}

func (s *UtilizationDataReader) GetInstanceList(ctx context.Context, instanceIdList ...string) ([]data.InstanceDetail, error) {
	// TODO 应该先判断是否有批量拿的接口
	result := make([]data.InstanceDetail, 0, len(instanceIdList))
	for _, i := range instanceIdList {
		resp, err := s._provider.DescribeInstanceAttribute(ctx, types.DescribeInstanceAttributeRequest{
			InstanceId: i,
		})
		if err != nil {
			return result, err
		}
		result = append(result, data.InstanceDetail{
			Provider:         s._provider.ProviderType(),
			InstanceId:       i,
			RegionId:         resp.RegionId,
			SubscriptionType: resp.SubscriptionType,
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