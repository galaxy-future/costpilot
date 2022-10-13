package tools

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

//
//import (
//	"fmt"
//	"reflect"
//	"testing"
//	"time"
//)
//
//func TestDate2Month(t *testing.T) {
//	type args struct {
//		date string
//	}
//	tests := []struct {
//		name string
//		args args
//		want string
//	}{
//		{
//			name: "test1",
//			args: args{
//				date: "2019-02-01",
//			},
//			want: "2019-02",
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := Date2Month(tt.args.date); got != tt.want {
//				t.Errorf("Date2Month() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestGetLastYearData(t *testing.T) {
//	type args struct {
//		data       []string
//		targetYear []string
//	}
//	tests := []struct {
//		name string
//		args args
//		want []string
//	}{
//		{
//			name: "test1",
//			args: args{
//				data:       []string{"2019-02-01", "2019-02-02"},
//				targetYear: nil,
//			},
//			want: []string{"2021-02-01", "2021-02-02"},
//		},
//		{
//			name: "test2",
//			args: args{
//				data:       []string{"2019-02-01", "2019-02-02"},
//				targetYear: []string{"2020"},
//			},
//			want: []string{"2020-02-01", "2020-02-02"},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := GetTargetYearData(tt.args.data, tt.args.targetYear...); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetTargetYearData() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
////
////func TestGetQuarterDate(t *testing.T) {
////	type args struct {
////		Month []string
////	}
////	tests := []struct {
////		name string
////		args args
////		want BillingDate
////	}{
////		{
////			name: "test1",
////			args: args{
////				//Month: "2022-09",
////			},
////			want: BillingDate{
////				Months: []string{"2022-09", "2022-10"},
////				Days: []string{"2022-10-01","2022-10-02","2022-10-03","2022-10-04","2022-10-05","2022-10-06","2022-10-07"},
////			},
////		},
////	}
////	for _, tt := range tests {
////		t.Run(tt.name, func(t *testing.T) {
////			if got := GetQuarterDate(); !reflect.DeepEqual(got, tt.want) {
////				t.Errorf("GetQuarterDate() = %v, want %v", got, tt.want)
////			}
////		})
////	}
////}
//
//func TestGetDaysInCurrentMonth(t *testing.T) {
//	tests := []struct {
//		name string
//		want []string
//	}{
//		{
//			name: "test1",
//			want: []string{
//				"2022-10-01",
//				"2022-10-02",
//				"2022-10-03",
//				"2022-10-04",
//				"2022-10-05",
//				"2022-10-06",
//				"2022-10-07",
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := GetDaysInRecentMonth(); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetDaysInRecentMonth() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestGetCurrentYearMonths(t *testing.T) {
//	tests := []struct {
//		name string
//		want []string
//	}{
//		{
//			name: "test1",
//			want: []string{
//				"2022-01",
//				"2022-02",
//				"2022-03",
//				"2022-04",
//				"2022-05",
//				"2022-06",
//				"2022-07",
//				"2022-08",
//				"2022-09",
//				"2022-10",
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := GetRecentYearMonths(); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetRecentYearMonths() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestGetQuarterDatPro(t *testing.T) {
//	type args struct {
//		isCurrentYear bool
//	}
//	tests := []struct {
//		name string
//		args args
//		want BillingDate
//	}{
//		{
//			name: "test-isCurrentYear_true",
//			args: args{
//				isCurrentYear: true,
//			},
//			want: BillingDate{
//				Months: []string{},
//				Days: []string{
//					"2021-10-01",
//					"2021-10-02",
//					"2021-10-03",
//					"2021-10-04",
//					"2021-10-05",
//					"2021-10-06",
//					"2021-10-07",
//				},
//			},
//		},
//		{
//			name: "test-isCurrentYear_false",
//			args: args{
//				isCurrentYear: false,
//			},
//			want: BillingDate{
//				Months: []string{},
//				Days: []string{
//					"2021-10-01",
//					"2021-10-02",
//					"2021-10-03",
//					"2021-10-04",
//					"2021-10-05",
//					"2021-10-06",
//					"2021-10-07",
//				},
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := GetRecentQuarterBillingDate(tt.args.isCurrentYear); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetRecentQuarterBillingDate() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestGetLastYear(t *testing.T) {
//	type args struct {
//		curYear []string
//	}
//	tests := []struct {
//		name string
//		args args
//		want string
//	}{
//		// TODO: Add test cases. liaoshengchen
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := GetLastYear(tt.args.curYear...); got != tt.want {
//				t.Errorf("GetPreviousYear() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//func TestGetLast14Days(t *testing.T) {
//	fmt.Println(GetRecentXDaysBillingDate(14))
//}
/*func TestGetLast12Months(t *testing.T) {
	p := NewBillDatePilot()
	p.SetNowT(time.Date(2022, 3, 12, 0, 0, 0, 0, time.Local))
	fmt.Println(p.GetRecentXMonthsBillingDate(5))
	fmt.Println(p.GetRecentXDaysBillingDate(5))
}*/

func TestConvBillingDate2LastMonth(t *testing.T) {
	p := NewBillDatePilot()
	days := make([]string, 0)
	for i := 1; i <= 30; i++ {
		days = append(days, "2022-03-"+fmt.Sprintf("%02d", i))
	}
	tests := []struct {
		name  string
		input BillingDate
		want  BillingDate
	}{
		{
			name:  "1",
			input: BillingDate{Days: days},
			want:  BillingDate{Months: []string{"2022-02"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret := p.ConvBillingDate2PreviousMonth(tt.input)
			fmt.Println(ret)
		})
	}
}
func TestConvBillingDate2LastQuarter(t *testing.T) {
	p := NewBillDatePilot()
	days := make([]string, 0)
	for i := 1; i <= 18; i++ {
		days = append(days, fmt.Sprintf("2022-05-%02d", i))
	}
	tests := []struct {
		name  string
		input BillingDate
		want  BillingDate
	}{
		{
			name: "1",
			input: BillingDate{
				Months: []string{"2022-04"},
				Days:   days,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret := p.ConvBillingDate2PreviousQuarter(tt.input)
			fmt.Println(ret)
		})
	}
}
func TestGetMonthBillingDate(t *testing.T) {
	p := NewBillDatePilot()
	tt, _ := time.Parse("2006-01-02", "2022-10-01")
	p.SetNowT(tt)
	fmt.Println(p.GetRecentMonthBillingDate(true))
}

func TestBillingDatePilot_GetYearBillingDate(t *testing.T) {
	type fields struct {
		_nowT time.Time
	}
	type args struct {
		isRecentYear bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   BillingDate
	}{
		// TODO: Add test cases.
		{
			name: "1",
			fields: fields{
				_nowT: time.Date(2022, 10, 10, 0, 0, 0, 0, time.Local),
			},
			args: args{
				isRecentYear: true,
			},
			want: BillingDate{Months: []string{"2022-01", "2022-02", "2022-03", "2022-04", "2022-05", "2022-06", "2022-07", "2022-08", "2022-09"}, Days: []string{"2022-10-01", "2022-10-02", "2022-10-03", "2022-10-04", "2022-10-05", "2022-10-06", "2022-10-07", "2022-10-08", "2022-10-09"}},
		},
		{
			name: "2",
			fields: fields{
				_nowT: time.Date(2022, 10, 10, 0, 0, 0, 0, time.Local),
			},
			args: args{
				isRecentYear: false,
			},
			want: BillingDate{Months: []string{"2021-01", "2021-02", "2021-03", "2021-04", "2021-05", "2021-06", "2021-07", "2021-08", "2021-09"}, Days: []string{"2021-10-01", "2021-10-02", "2021-10-03", "2021-10-04", "2021-10-05", "2021-10-06", "2021-10-07", "2021-10-08", "2021-10-09"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &BillingDatePilot{
				_nowT: tt.fields._nowT,
			}
			if tt.args.isRecentYear {
				if got := p.GetRecentYearBillingDate(); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetRecentYearBillingDate() = %v, want %v", got, tt.want)
				}
			} else {
				if got := p.GetPreviousYearBillingDate(); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetPreviousYearBillingDate() = %v, want %v", got, tt.want)
				}
			}

		})
	}
}
func TestGetLastXMonthsBillingDate(t *testing.T) {
	p := NewBillDatePilot()
	tt, _ := time.Parse("2006-01-02", "2022-10-10")
	p.SetNowT(tt)
	fmt.Println(p.GetRecentXMonthsBillingDate(2))
}
