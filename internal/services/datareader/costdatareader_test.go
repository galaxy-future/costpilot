package datareader

import (
	"context"
	"testing"

	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/data"
	"github.com/galaxy-future/costpilot/internal/providers"
	jsoniter "github.com/json-iterator/go"
)

var _AK = "ak_test_123"
var _SK = "sk_test_123"

func TestCostDataRader_GetDailyCost(t *testing.T) {
	type fields struct {
		_provider providers.Provider
	}
	type args struct {
		ctx              context.Context
		date             string
		isGroupByProduct bool
	}
	provider, err := providers.GetProvider(cloud.AlibabaCloud, _AK, _SK, "cn-beijing")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    data.DailyBilling
		wantErr bool
	}{
		{
			name: "GetCostByDate-Group_true",
			fields: fields{
				_provider: provider,
			},
			args: args{
				ctx:              context.Background(),
				date:             "2022-07-15",
				isGroupByProduct: true,
			},
			want:    data.DailyBilling{},
			wantErr: false,
		},
		{
			name: "GetCostByDate-Group_false",
			fields: fields{
				_provider: provider,
			},
			args: args{
				date:             "2022-07-15",
				isGroupByProduct: false,
			},
			want:    data.DailyBilling{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &CostDataReader{
				_provider: tt.fields._provider,
			}
			got, err := s.GetDailyCost(tt.args.ctx, tt.args.date, tt.args.isGroupByProduct)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDailyCost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			str, _ := jsoniter.MarshalToString(got)
			t.Logf("QueryAccountBill() got = %s", str)
		})
	}
}

func TestCostDataRader_GetMonthlyCost(t *testing.T) {
	type fields struct {
		_provider providers.Provider
	}
	type args struct {
		ctx              context.Context
		month            string
		isGroupByProduct bool
	}
	provider, err := providers.GetProvider(cloud.AlibabaCloud, _AK, _SK, "cn-beijing")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    data.DailyBilling
		wantErr bool
	}{
		{
			name: "GetCostByMont-Group_true",
			fields: fields{
				_provider: provider,
			},
			args: args{
				ctx:              context.Background(),
				month:            "2022-08",
				isGroupByProduct: true,
			},
			want:    data.DailyBilling{},
			wantErr: false,
		},
		{
			name: "GetCostByMonth-Group_false",
			fields: fields{
				_provider: provider,
			},
			args: args{
				month:            "2022-08",
				isGroupByProduct: false,
			},
			want:    data.DailyBilling{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &CostDataReader{
				_provider: tt.fields._provider,
			}
			got, err := s.GetMonthlyCost(tt.args.ctx, tt.args.month, tt.args.isGroupByProduct)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDailyCost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			str, _ := jsoniter.MarshalToString(got)
			t.Logf("QueryAccountBill() got = %s", str)
		})
	}
}
