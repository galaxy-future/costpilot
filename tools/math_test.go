package tools

import (
	"fmt"
	"testing"

	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/data"
	"github.com/galaxy-future/costpilot/internal/providers/types"
)

func TestFloat64Add(t *testing.T) {
	type args struct {
		a []float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "test1",
			args: args{
				a: []float64{0.1, 0.23},
			},
			want: 0.33,
		},
		{
			name: "test2",
			args: args{
				a: []float64{},
			},
			want: 0,
		},
		{
			name: "test3",
			args: args{
				a: []float64{0.25},
			},
			want: 0.25,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Float64Add(tt.args.a...); got != tt.want {
				t.Errorf("Float64Add() = %v, want %v", got, tt.want)
			}
		})
	}

}
func TestAddProductBilling(t *testing.T) {
	x := data.ProductBilling{
		ProductName: types.ECS.String(),
		TotalAmount: 10,
		Items: []data.ItemInProductBilling{
			{
				PipCode:          types.ECS.String(),
				ProductName:      types.ECS.String(),
				PretaxAmount:     6,
				SubscriptionType: cloud.PrePaid,
			},
			{
				PipCode:          types.ECS.String(),
				ProductName:      types.ECS.String(),
				PretaxAmount:     4,
				SubscriptionType: cloud.PostPaid,
			},
		},
	}
	y := data.ProductBilling{
		ProductName: types.ECS.String(),
		TotalAmount: 60,
		Items: []data.ItemInProductBilling{
			{
				PipCode:          types.ECS.String(),
				ProductName:      types.ECS.String(),
				PretaxAmount:     60,
				SubscriptionType: cloud.PrePaid,
			},
			/*			{
						PipCode:          types.ECS.String(),
						ProductName:      types.ECS.String(),
						PretaxAmount:     40,
						SubscriptionType: cloud.PostPaid.String(),
					},*/
		},
	}
	fmt.Println(AddProductBilling(x, y))

}
