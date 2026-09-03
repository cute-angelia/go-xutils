package itime

import (
	"testing"
	"time"
)

func TestToUnix(t *testing.T) {
	tm1 := NewUnixNow()
	// 输出: 2026-01-01
	t.Log(tm1.Format(WithSetCST(), WithSetStartOfMonth(), WithFormatOnlyDate()))

	// CST
	t.Log(tm1.Format(WithSetCST()))

	t.Log(tm1.Format(WithSetEndOfDay()))

	t.Log(tm1.Format(WithSetEndOfWeek(), WithFormatOnlyDate()))
}

func TestAddDateUpdatesTheTime(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	start := time.Date(2026, 8, 29, 10, 11, 12, 0, loc)
	tm := NewTime(start)

	got := tm.AddDate(0, 0, 30)

	wantTime := start.AddDate(0, 0, 30)
	if got.GetUnix() != wantTime.Unix() {
		t.Fatalf("AddDate unix = %d, want %d", got.GetUnix(), wantTime.Unix())
	}

	wantFormatted := "2026-09-28 10:11:12"
	if got.Format(WithSetCST()) != wantFormatted {
		t.Fatalf("AddDate format = %q, want %q", got.Format(WithSetCST()), wantFormatted)
	}
}

func TestGetBetweenDatesWithSameDateReturnsSingleDay(t *testing.T) {
	done := make(chan []string, 1)
	go func() {
		done <- GetBetweenDates("2020-01-01", "2020-01-01")
	}()

	select {
	case got := <-done:
		want := []string{"2020-01-01"}
		if len(got) != len(want) || got[0] != want[0] {
			t.Fatalf("GetBetweenDates() = %#v, want %#v", got, want)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("GetBetweenDates did not return for identical start and end dates")
	}
}

func TestGetBetweenDatesWithSameCalendarDateAndDifferentTimesReturnsSingleDay(t *testing.T) {
	got := GetBetweenDates("2020-01-01 10:00:00", "2020-01-01 12:00:00")
	want := []string{"2020-01-01"}

	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("GetBetweenDates() = %#v, want %#v", got, want)
	}
}

func TestDateRangeEndTimesIncludeEntireFinalDay(t *testing.T) {
	_, weekEnd := GetWeekDate()
	_, monthEnd := GetMonthDate()
	_, yearEnd := GetYearDate()

	if got := weekEnd.Format(WithFormatOnlyTime()); got != "23:59:59" {
		t.Fatalf("GetWeekDate end time = %q, want 23:59:59", got)
	}
	if got := monthEnd.Format(WithFormatOnlyTime()); got != "23:59:59" {
		t.Fatalf("GetMonthDate end time = %q, want 23:59:59", got)
	}
	if got := yearEnd.Format(WithFormatOnlyTime()); got != "23:59:59" {
		t.Fatalf("GetYearDate end time = %q, want 23:59:59", got)
	}
}
