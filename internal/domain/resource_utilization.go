package domain

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/galayx-future/costpilot/internal/services"
	"github.com/galayx-future/costpilot/internal/services/databean"
	"github.com/galayx-future/costpilot/internal/services/template"
	"github.com/galayx-future/costpilot/internal/types"
)

type ResourceUtilizationDomain struct {
	nowT                     time.Time
	dailyCpuProviders        []*sync.Map
	dailyMemoryProviders     []*sync.Map
	recentInstancesProviders []*sync.Map
}

func NewResourceUtilizationDomain() *ResourceUtilizationDomain {
	return &ResourceUtilizationDomain{
		nowT:                     time.Now(),
		dailyCpuProviders:        []*sync.Map{},
		dailyMemoryProviders:     []*sync.Map{},
		recentInstancesProviders: []*sync.Map{},
	}
}

// GetUtilization 获取资源利用情况
func (s *ResourceUtilizationDomain) GetUtilization(ctx context.Context, a types.CloudAccount) (dailyCpu, dailyMemory, dailyInstances *sync.Map, err error) {
	dBean := databean.NewUtilization(a, s.nowT)
	err = dBean.RunPipeline(ctx)
	if err != nil {
		return
	}
	dailyCpu, dailyMemory, dailyInstances = dBean.GetUtilizationMap()
	return
}

func (s *ResourceUtilizationDomain) GetUtilizationData(ctx context.Context) error {
	accounts := services.NewAccountService().GetAccounts()
	if len(accounts) == 0 {
		log.Println("E! cloud account is not configured, please check conf/config.yml")
		return errors.New("cloud account is not configured")
	}
	for _, a := range accounts {
		dailyCpu, dailyMemory, dailyInstances, err := s.GetUtilization(ctx, a)
		if err != nil {
			log.Printf("E! get cloud-acount[%v] utilization error = %v", a.Name, err)
			return err
		}
		s.dailyMemoryProviders = append(s.dailyCpuProviders, dailyMemory)
		s.dailyCpuProviders = append(s.dailyCpuProviders, dailyCpu)
		s.recentInstancesProviders = append(s.recentInstancesProviders, dailyInstances)
	}
	return nil
}

func (s *ResourceUtilizationDomain) ExportStatisticData(ctx context.Context) error {
	temp := template.NewUtilization(s.nowT)
	temp.AssignData(s.dailyCpuProviders, s.dailyMemoryProviders, s.recentInstancesProviders)
	data := temp.Assemble(ctx)
	err := temp.Export(ctx, data)
	if err != nil {
		log.Printf("E! export cost-analysis data failed: %v\n", err)
		return err
	}

	return nil
}

func (s *ResourceUtilizationDomain) GetCostAnalysisPipeline() []func(context.Context) error {
	return []func(context.Context) error{
		s.GetUtilizationData,
		s.ExportStatisticData,
	}
}

func (s *ResourceUtilizationDomain) RunPipeline(ctx context.Context) error {
	var err error
	for _, f := range s.GetCostAnalysisPipeline() {
		err = f(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
