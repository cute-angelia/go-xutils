package itime

import (
	"testing"
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
