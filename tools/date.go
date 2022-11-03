package tools

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"
)

type BillingDatePilot struct {
	_nowT time.Time
}

func NewBillDatePilot() *BillingDatePilot {
	return &BillingDatePilot{
		_nowT: time.Now(),
	}
}

func (p *BillingDatePilot) SetNowT(t time.Time) *BillingDatePilot {
	p._nowT = t
	return p
}

func (p *BillingDatePilot) GetNowT() time.Time {
	return p._nowT
}

type BillingDate struct {
	Months []string `json:"months"` //["2022-02"]
	Days   []string `json:"days"`   //["2022-03-01","2022-03-02", ..."2022-03-29"]
}

// Date2Month 2022-10-09 -> 2022-10
func Date2Month(date string) string {
	s := strings.Split(date, "-")
	return fmt.Sprintf("%s-%s", s[0], s[1])
}

// GetYearMonths
// 获取今年截至今天的所有月份 [2022-01,2022-02, ..., 2022-10]
func (p *BillingDatePilot) GetRecentYearMonths() []string {
	ret := []string{}
	monthOneS := p._nowT.Format("2006") + "-01"
	currentMonthS := p._nowT.Format("2006-01")
	start, _ := time.ParseInLocation("2006-01", monthOneS, time.Local)
	end, _ := time.ParseInLocation("2006-01", currentMonthS, time.Local)
	for i := start; i.Before(end); {
		ret = append(ret, i.Format("2006-01"))
		i = i.AddDate(0, 1, 0)
	}
	ret = append(ret, end.Format("2006-01"))
	return ret
}

// GetDaysInRecentMonth
// 获取截至今天的当前月份的日期 [2022-10-01, 2022-10-02, ..., 2022-10-07]
func (p *BillingDatePilot) GetDaysInRecentMonth() []string {
	ret := []string{}
	dayOneS := p._nowT.Format("2006-01") + "-01"
	todayS := p._nowT.Format("2006-01-02")
	start, _ := time.ParseInLocation("2006-01-02", dayOneS, time.Local)
	end, _ := time.ParseInLocation("2006-01-02", todayS, time.Local)
	for i := start; i.Before(end); {
		ret = append(ret, i.Format("2006-01-02"))
		i = i.AddDate(0, 0, 1)
	}
	return ret
}

// ConvBillingDate2PreviousMonth
func (p *BillingDatePilot) ConvBillingDate2PreviousMonth(d BillingDate) BillingDate {
	ret := BillingDate{
		Months: make([]string, 0),
		Days:   make([]string, 0),
	}
	for _, m := range d.Months {
		t, _ := time.Parse("2006-01", m)
		ret.Months = append(ret.Months, t.AddDate(0, -1, 0).Format("2006-01"))
	}
	if len(d.Days) != 0 {
		firstDay, _ := time.Parse("2006-01-02", d.Days[0])
		lastMonthFirstDay := AddDate(firstDay, 0, -1, 0)
		if lastMonthFirstDay.AddDate(0, 0, len(d.Days)).Month() == firstDay.Month() {
			ret.Months = append(ret.Months, lastMonthFirstDay.Format("2006-01"))
			return ret
		}
		for _, s := range d.Days {
			t, _ := time.Parse("2006-01-02", s)
			ret.Days = append(ret.Days, t.AddDate(0, -1, 0).Format("2006-01-02"))
		}
	}

	return ret
}

// ConvDays2LastQuarter
func (p *BillingDatePilot) ConvBillingDate2PreviousQuarter(d BillingDate) BillingDate {
	ret := BillingDate{
		Months: make([]string, 0),
		Days:   make([]string, 0),
	}
	for _, s := range d.Months {
		t, _ := time.Parse("2006-01", s)
		ret.Months = append(ret.Months, t.AddDate(0, -3, 0).Format("2006-01"))
	}
	if len(d.Days) != 0 {
		firstDay, _ := time.Parse("2006-01-02", d.Days[0])
		lastQuarterFirstDay := AddDate(firstDay, 0, -3, 0)
		if lastQuarterFirstDay.AddDate(0, 0, len(d.Days)).Month() == AddDate(firstDay, 0, -2, 0).Month() {
			ret.Months = append(ret.Months, lastQuarterFirstDay.Format("2006-01"))
			return ret
		}
		for _, s := range d.Days {
			t, _ := time.Parse("2006-01-02", s)
			ret.Days = append(ret.Days, t.AddDate(0, -3, 0).Format("2006-01-02"))
		}
	}

	return ret
}

// GetRecentMonth
func (p *BillingDatePilot) GetRecentMonth() string {
	if p._nowT.Day() == 1 {
		return AddDate(p._nowT, 0, 0, -1).Format("2006-01") // TODO liaoshengchen 此处可有 bug
	}
	return p._nowT.Format("2006-01")
}

// GetPreviousMonth
func (p *BillingDatePilot) GetPreviousMonth() string {
	m := p.GetRecentMonth()
	t, _ := time.ParseInLocation("2006-01", m, time.Local)
	return AddDate(t, 0, -1, 0).Format("2006-01") // TODO liaoshengchen 此处可有 bug
}

// GetPreviousYear
func (p *BillingDatePilot) GetPreviousYear(curYear ...string) string {
	t := p._nowT
	if len(curYear) > 0 {
		t, _ = time.ParseInLocation("2006", curYear[0], time.Local)
	}
	return fmt.Sprintf("%d", cast.ToInt64(t.Format("2006"))-1)
}

// GetTargetYearData 获取目标年的年/月/日数据
// offset : negative value for the year before ,vice versa
func (p *BillingDatePilot) GetTargetYearData(data []string, offset int) []string {
	if len(data) == 0 {
		return []string{}
	}
	for i, date := range data {
		num, _ := strconv.Atoi(date[0:4])
		data[i] = strconv.Itoa(num+offset) + date[4:]
	}
	return data
}

// isFirstDayOfMonth
func (p *BillingDatePilot) isFirstDayOfMonth() bool {
	return p._nowT.Day() == 1
}

// GetRecentYearBillingDate
func (p *BillingDatePilot) GetRecentYearBillingDate() BillingDate {
	return p.GetBillingDate(true, []int64{1})
}

// GetPreviousYearBillingDate
func (p *BillingDatePilot) GetPreviousYearBillingDate() BillingDate {
	return p.GetBillingDate(false, []int64{1})
}

// GetRecentQuarterBillingDate
func (p *BillingDatePilot) GetRecentQuarterBillingDate(isRecentYear bool) BillingDate {
	return p.GetBillingDate(isRecentYear, []int64{10, 7, 4, 1}) //first month in quarter
}

// GetRecentMonthBillingDate
func (p *BillingDatePilot) GetRecentMonthBillingDate(isRecentYear bool) BillingDate {
	return p.GetBillingDate(isRecentYear, []int64{12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1})
}

// GetBillingDate isCurrentYear: true 今年 | false 去年
func (p *BillingDatePilot) GetBillingDate(isRecentYear bool, firstMonths []int64) BillingDate {
	ret := BillingDate{
		Months: []string{},
		Days:   p.GetDaysInRecentMonth(),
	}
	var idxMonth string

	if p.isFirstDayOfMonth() {
		idxMonth = p.GetRecentMonth()
	} else {
		idxMonth = p._nowT.Format("2006-01")
	}
	year := idxMonth[0:4]
	m := cast.ToInt64(strings.TrimLeft(strings.Split(idxMonth, "-")[1], "0"))
	for _, v := range firstMonths {
		if m >= v {
			i := v
			for m >= i {
				month := fmt.Sprintf("%s-%02d", year, i)
				ret.Months = append(ret.Months, month)
				i++
			}
			break
		}
	}
	if len(ret.Days) != 0 { //当月天数不为 0
		ret.Months = ret.Months[0 : len(ret.Months)-1]
	}
	if isRecentYear {
		return ret
	}
	ret.Months = p.GetTargetYearData(ret.Months, -1)
	ret.Days = p.GetTargetYearData(ret.Days, -1)

	return ret
}

// GetRecentXDaysBillingDate 返回过去 X 天(不含今天)的日期 ["2022-09-24","2022-09-25", ... "2022-10-07"]
func (p *BillingDatePilot) GetRecentXDaysBillingDate(x int32) BillingDate { // TODO liaoshengchen
	t := p._nowT
	ret := make([]string, 0, 14)
	for i := int(x); i > 0; i-- {
		tt := t.AddDate(0, 0, -i)
		ret = append(ret, tt.Format("2006-01-02"))
	}
	return BillingDate{
		Days: ret,
	}
}

// GetRecentXMonthsBillingDate
// 返回过去 x 个月(不含 1 号当月) ["2021-11","2021-12", ... "2022-09",]
func (p *BillingDatePilot) GetRecentXMonthsBillingDate(x int32) BillingDate {
	t := p._nowT
	days := make([]string, 0)
	for i := 1; i < t.Day(); i++ {
		days = append(days, fmt.Sprintf(t.Format("2006-01")+"-%02d", i))
	}
	ret := make([]string, 0, x+1)
	if p.isFirstDayOfMonth() {
		x++
	}
	for i := int(x - 1); i >= 1; i-- {
		tt := AddDate(t, 0, -i, 0)
		ret = append(ret, tt.Format("2006-01"))
	}
	return BillingDate{
		Months: ret,
		Days:   days,
	}
}

// IsFirstDayOfYear
func (p *BillingDatePilot) IsFirstDayOfYear() bool {
	return p._nowT.Format("01-02") == "01-01"
}

// GetRecentDayBillingDate
func (p *BillingDatePilot) GetRecentDayBillingDate() BillingDate {
	day := p._nowT.AddDate(0, 0, -1).Format("2006-01-02")
	return BillingDate{
		Days: []string{day},
	}
}

// GetPreviousDayBillingDate
func (p *BillingDatePilot) GetPreviousDayBillingDate() BillingDate {
	day := p._nowT.AddDate(0, 0, -2).Format("2006-01-02")
	return BillingDate{
		Days: []string{day},
	}
}

// GetRecentQuarter
func (p *BillingDatePilot) GetRecentQuarter() int {
	month := int(p._nowT.AddDate(0, 0, -1).Month())
	return (month-1)/3 + 1
}

// GetRecentYear
func (p *BillingDatePilot) GetRecentYear() int {
	return p._nowT.AddDate(0, 0, -1).Year()
}

// IsValidDayDate
// 2006-01-02 满足格式且真实存在的日期
func IsValidDayDate(d string) bool {
	if _, err := time.Parse("2006-01-02", d); err != nil {
		return false
	}
	return true
}

// IsValidMonthDate
// 2006-01 满足格式且真实存在的日期
func IsValidMonthDate(d string) bool {
	if _, err := time.Parse("2006-01", d); err != nil {
		return false
	}
	return true
}
