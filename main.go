package main

import (
	"context"
	"os"

	_ "github.com/galaxy-future/costpilot/tools"

	"github.com/galaxy-future/costpilot/internal/domain"

	"github.com/galaxy-future/costpilot/internal/config"
)

func main() {
	ctx := context.Background()
	printVersion()
	if err := config.Init(); err != nil {
		os.Exit(1)
	}

	a := domain.NewCostAnalysisDomain()
	if err := a.RunPipeline(ctx); err != nil {
		os.Exit(1)
	}

	b := domain.NewResourceUtilizationDomain()
	if err := b.RunPipeline(ctx); err != nil {
		os.Exit(1)
	}

	if err := output(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
