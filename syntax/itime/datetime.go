package itime

import (
	"log"
	"strings"
	"time"
)

// @referer https://github.com/duke-git/lancet/blob/main/datetime/conversion.go

type TheTime struct {
	unix int64
}

// 预定义时区，避免重复创建
var (
	shanghaiLoc = time.FixedZone("CST", 8*3600)
)

// FormatOption ======================================== Functional Options ========================================
// Option 定義函數特徵
type FormatOption func(tm *time.Time, layout *string)

// WithSetCST 选项：转换为CST China Standard Time (UTC+8)
func WithSetCST() FormatOption {
	return func(tm *time.Time, layout *string) {
		*tm = tm.In(shanghaiLoc)
	}
}

// WithFormatLayout 自定義格式
func WithFormatLayout(newLayout string) FormatOption {
	return func(tm *time.Time, layout *string) {
		*layout = newLayout
	}
}

func WithFormatOnlyYear() FormatOption {
	return WithFormatLayout("2006")
}

// WithFormatOnlyDate 預定義的快捷格式選項：僅輸出日期部分 (2006-01-02)
func WithFormatOnlyDate() FormatOption {
	return WithFormatLayout("2006-01-02")
}

// WithFormatOnlyTime 預定義的快捷格式選項：僅輸出時間部分 (15:04:05)
func WithFormatOnlyTime() FormatOption {
	return WithFormatLayout("15:04:05")
}

// WithSetStartOfDay 将时间设置为当天的 00:00:00
func WithSetStartOfDay() FormatOption {
	return func(tm *time.Time, layout *string) {
		// 必须先处理时区，再计算零点，否则跨时区转换会导致日期跳变
		y, m, d := tm.Date()
		*tm = time.Date(y, m, d, 0, 0, 0, 0, tm.Location())
	}
}

// WithSetEndOfDay 将时间设置为当天的 23:59:59
func WithSetEndOfDay() FormatOption {
	return func(tm *time.Time, layout *string) {
		y, m, d := tm.Date()
		*tm = time.Date(y, m, d, 23, 59, 59, 0, tm.Location())
	}
}

// WithSetISO8601 切换格式为 RFC3339 (ISO8601 标准)
func WithSetISO8601() FormatOption {
	return func(tm *time.Time, layout *string) {
		*layout = time.RFC3339
	}
}

// WithSetStartOfWeek 将时间调整为本周一的 00:00:00
func WithSetStartOfWeek() FormatOption {
	return func(tm *time.Time, layout *string) {
		loc := tm.Location()

		// Go 的 Weekday: Sunday=0, Monday=1, ..., Saturday=6
		weekday := int(tm.Weekday())

		// 计算距离本周一的差值
		// 如果是周日(0)，距离周一就是 -6 天
		// 如果是周一(1)，距离周一就是 0 天
		// 如果是周二(2)，距离周一就是 -1 天
		offset := 1 - weekday
		if weekday == 0 {
			offset = -6
		}

		// 重新构造该时区下的周一零点
		*tm = time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, loc).
			AddDate(0, 0, offset)
	}
}

// WithSetEndOfWeek 获得当前周的初始和结束日期
func WithSetEndOfWeek() FormatOption {
	return func(tm *time.Time, layout *string) {
		// 1. 获取当前 tm 的时区（可能是 Local，也可能是 WithCST 设定的）
		loc := tm.Location()

		// 2. 修正周日为 7 的逻辑
		weekday := int(tm.Weekday())
		if weekday == 0 {
			weekday = 7
		}

		// 3. 计算距离周日的差值
		daysToSunday := 7 - weekday

		// 4. 重新构造该时区下的周日最后一秒
		*tm = time.Date(tm.Year(), tm.Month(), tm.Day(), 23, 59, 59, 0, loc).
			AddDate(0, 0, daysToSunday)
	}
}

// WithSetStartOfMonth 将时间调整为本月 1 号的 00:00:00
func WithSetStartOfMonth() FormatOption {
	return func(tm *time.Time, layout *string) {
		y, m, _ := tm.Date()
		// 直接构造本月 1 号
		*tm = time.Date(y, m, 1, 0, 0, 0, 0, tm.Location())
	}
}

// WithSetEndOfMonth 将时间调整为本月最后一天的 23:59:59
func WithSetEndOfMonth() FormatOption {
	return func(tm *time.Time, layout *string) {
		y, m, _ := tm.Date()
		// 逻辑：下个月的 1 号，减去 1 秒（或者 AddDate(0, 1, -1) 后设为 23:59:59）
		firstOfNextMonth := time.Date(y, m+1, 1, 0, 0, 0, 0, tm.Location())
		*tm = firstOfNextMonth.Add(-time.Second)
	}
}

// ======================================== Functional Options End ========================================

// NewUnixNow return unix timestamp of current time
func NewUnixNow() *TheTime {
	return &TheTime{unix: time.Now().Unix()}
}

func NewTime(t time.Time) *TheTime {
	return NewUnix(t.Unix())
}

// NewUnix return unix timestamp of specified time
func NewUnix(unix int64) *TheTime {
	return &TheTime{unix: unix}
}

// NewFormat return unix timestamp of specified time string, t should be "yyyy-mm-dd hh:mm:ss"
func NewFormat(t string) (*TheTime, error) {
	timeLayout := "2006-01-02 15:04:05"

	// 包含 T
	if strings.Contains(t, "T") {
		timeLayout = time.RFC3339
	}

	loc := time.FixedZone("CST", 8*3600)
	tt, err := time.ParseInLocation(timeLayout, t, loc)
	if err != nil {
		return nil, err
	}
	return &TheTime{unix: tt.Unix()}, nil
}

// NewFormatLayout 根据模板获取时间 强制+8时区
func NewFormatLayout(t string, timeLayout string) *TheTime {
	// 包含 T
	if strings.Contains(t, "T") {
		timeLayout = time.RFC3339
	}
	loc := time.FixedZone("CST", 8*3600)
	tt, err := time.ParseInLocation(timeLayout, t, loc)
	if err != nil {
		log.Println("NewFormatLayout error", t, timeLayout, err)
	}
	return &TheTime{unix: tt.Unix()}
}

// NewISO8601 return unix timestamp of specified iso8601 time string
func NewISO8601(iso8601 string) (*TheTime, error) {
	t, err := time.ParseInLocation(time.RFC3339, iso8601, time.UTC)
	if err != nil {
		return nil, err
	}
	return &TheTime{unix: t.Unix()}, nil
}

// GetTime return time.Time
func (t *TheTime) GetTime() time.Time {
	return time.Unix(t.unix, 0)
}

// GetUnix return unix timestamp
func (t *TheTime) GetUnix() int64 {
	return t.unix
}

// GetMillisecond 获取毫秒
func (t *TheTime) GetMillisecond() int64 {
	return t.GetTime().UnixNano() / int64(time.Millisecond)
}

// GetWeekDayChinese 中国的周码 （1，2，3，4，5，6，7）
func (t *TheTime) GetWeekDayChinese() int {
	now := t.GetTime()
	if now.Weekday() == 0 {
		return 7
	} else {
		return int(now.Weekday())
	}
}

func (t *TheTime) AddDate(years int, months int, days int) *TheTime {
	t.GetTime().AddDate(years, months, days)
	return t
}

// --------------------- format -----------------------

// Format 格式化时间字符串
func (t *TheTime) Format(opts ...FormatOption) string {
	if t == nil {
		return ""
	}
	// 1. 初始狀態：預設格式與本地時間
	tm := time.Unix(t.unix, 0)
	layout := "2006-01-02 15:04:05"

	// 2. 執行所有選項操作
	for _, opt := range opts {
		opt(&tm, &layout)
	}
	return tm.Format(layout)
}
