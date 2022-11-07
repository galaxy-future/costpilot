package datareader

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/galaxy-future/costpilot/internal/providers/alibaba"
)

func TestUtilizationDataReader_GetAllRegionMap(t *testing.T) {
	p, _ := alibaba.New(_AK, _SK, "")
	s := &UtilizationDataReader{
		_provider: p,
	}
	got, err := s.GetAllRegionMap(context.TODO())
	if err != nil {
		t.Errorf("GetAllRegionMap() error = %v", err)
		return
	}
	assert.NotEmpty(t, got)
	assert.NotEmpty(t, len(got))
}
