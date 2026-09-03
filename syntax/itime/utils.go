package itime

import (
	"fmt"
	"time"
)

// GetMonthDate 获得当前月的初始和结束日期
func GetMonthDate() (*TheTime, *TheTime) {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, 0).Add(-time.Second)
	return &TheTime{unix: firstOfMonth.Unix()}, &TheTime{unix: lastOfMonth.Unix()}
}

// GetWeekDate 获得当前周的初始和结束日期
func GetWeekDate() (*TheTime, *TheTime) {
	now := time.Now()
	offset := int(time.Monday - now.Weekday())
	//周日做特殊判断 因为time.Monday = 0
	if offset > 0 {
		offset = -6
	}
	lastoffset := int(time.Saturday - now.Weekday())
	//周日做特殊判断 因为time.Monday = 0
	if lastoffset == 6 {
		lastoffset = -1
	}
	firstOfWeek := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	lastOfWeeK := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, lastoffset+2).Add(-time.Second)
	return &TheTime{unix: firstOfWeek.Unix()}, &TheTime{unix: lastOfWeeK.Unix()}
}

// GetQuarterDate 获得当前季度的初始和结束日期
func GetQuarterDate() (*TheTime, *TheTime, int) {
	year := time.Now().Format("2006")
	month := int(time.Now().Month())
	var firstOfQuarter string
	var lastOfQuarter string
	var quarter int

	if month >= 1 && month <= 3 {
		//1月1号
		firstOfQuarter = year + "-01-01 00:00:00"
		lastOfQuarter = year + "-03-31 23:59:59"
		quarter = 1
	} else if month >= 4 && month <= 6 {
		firstOfQuarter = year + "-04-01 00:00:00"
		lastOfQuarter = year + "-06-30 23:59:59"
		quarter = 2
	} else if month >= 7 && month <= 9 {
		firstOfQuarter = year + "-07-01 00:00:00"
		lastOfQuarter = year + "-09-30 23:59:59"
		quarter = 3
	} else {
		firstOfQuarter = year + "-10-01 00:00:00"
		lastOfQuarter = year + "-12-31 23:59:59"
		quarter = 4
	}
	return NewFormatLayout(firstOfQuarter, TIME_FORMAT), NewFormatLayout(lastOfQuarter, TIME_FORMAT), quarter
}

func GetYearDate() (*TheTime, *TheTime) {
	now := time.Now()
	currentYear, _, _ := now.Date()
	currentLocation := now.Location()

	first := time.Date(currentYear, 1, 1, 0, 0, 0, 0, currentLocation)
	last := first.AddDate(1, 0, 0).Add(-time.Second)
	return &TheTime{unix: first.Unix()}, &TheTime{unix: last.Unix()}
}

// GetBetweenDates 根据开始日期和结束日期计算出时间段内所有日期
// 参数为日期格式，如：2020-01-01
func GetBetweenDates(sdate, edate string) []string {
	d := []string{}
	timeFormatTpl := "2006-01-02 15:04:05"
	if len(sdate) > len(timeFormatTpl) || len(edate) > len(timeFormatTpl) {
		return d
	}
	if len(timeFormatTpl) != len(sdate) {
		timeFormatTpl = timeFormatTpl[0:len(sdate)]
	}
	date, err := time.Parse(timeFormatTpl, sdate)
	if err != nil {
		// 时间解析，异常
		return d
	}
	date2, err := time.Parse(timeFormatTpl, edate)
	if err != nil {
		// 时间解析，异常
		return d
	}
	if date2.Before(date) {
		// 如果结束时间小于开始时间，异常
		return d
	}
	// 输出日期格式固定
	timeFormatTpl = "2006-01-02"
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	date2 = time.Date(date2.Year(), date2.Month(), date2.Day(), 0, 0, 0, 0, date2.Location())
	for !date.After(date2) {
		d = append(d, date.Format(timeFormatTpl))
		date = date.AddDate(0, 0, 1)
	}
	return d
}

// ConvertVideoSecToStr 视频时长多少秒 =》 转 slice
// 100s => []string{00,01,40}
func ConvertVideoSecToStr(sec int64) []string {
	if sec <= 0 {
		return []string{}
	}
	lastsec := sec
	h := sec / 3600
	if h > 0 {
		lastsec = lastsec - h*3600
	}
	m := lastsec / 60
	if m > 0 {
		lastsec = lastsec - m*60
	}
	s := lastsec
	return []string{fmt.Sprintf("%02d", h), fmt.Sprintf("%02d", m), fmt.Sprintf("%02d", s)}
}
