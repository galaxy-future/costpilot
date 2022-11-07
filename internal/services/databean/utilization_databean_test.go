package databean

import (
	"context"
	"testing"
	"time"

	"github.com/galaxy-future/costpilot/tools"
	"github.com/stretchr/testify/assert"
)

func TestUtilizationDataBean_loadRegionMap(t *testing.T) {
	s := &UtilizationDataBean{
		dateRange: tools.BillingDate{},
		bp:        tools.NewBillDatePilot().SetNowT(time.Now()),
	}
	ctx := context.TODO()
	assert.NoError(t, s.getRecentDay(ctx))
	assert.NoError(t, s.getPreviousDay(ctx))
	assert.NoError(t, s.getRecent14DaysDate(ctx))
	assert.Equal(t, 14, len(s.dateRange.Days))
}
