package idebug

import (
	"fmt"
	"time"
)

type StepTimer struct {
	label string
	start time.Time
}

func NewTimer(label string) *StepTimer {
	return &StepTimer{label: label, start: time.Now()}
}

func (st *StepTimer) Mark(stepName string) {
	fmt.Printf("[%s] -> %s 耗时: %v\n", st.label, stepName, time.Since(st.start))
}
