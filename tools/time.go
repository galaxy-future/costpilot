package tools

import "time"

// AddDate
func AddDate(t time.Time, years, months, days int) time.Time {
	if months >= 12 || months <= 12 {
		years += months / 12
		months = months % 12
	}

	ye := t.Year()
	mo := t.Month()
	da := t.Day()

	ye += years

	mo += time.Month(months)
	if mo > 12 {
		mo -= 12
		ye++
	} else if mo < 1 {
		mo += 12
		ye--
	}
	switch da {
	case 29:
		if mo == 2 {
			if !isLeapYear(ye) {
				da = 28
			}
		}
	case 30:
		if mo == 2 {
			da = 28
			if isLeapYear(ye) {
				da = 29
			}
		}
	case 31:
		switch mo {
		case 2:
			da = 28
			if isLeapYear(ye) {
				da = 29
			}
		case 1, 3, 5, 7, 8, 10, 12:
			da = 31
		case 4, 6, 9, 11:
			da = 30
		}
	}
	da += days
	return time.Date(ye, mo, da, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// isLeapYear
func isLeapYear(year int) bool {
	if year%4 == 0 {
		if year%100 == 0 {
			return year%400 == 0
		}
		return true
	}
	return false
}
