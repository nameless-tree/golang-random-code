package extra

import "time"

func UNUSED(C ...interface{}) {}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

func IsTimeRangeWithin(target, check TimeRange) bool {
	return target.Start.Equal(check.Start) || (target.Start.After(check.Start) && target.End.Before(check.End))
}

func Time12Parse(timeStr string) (time.Time, error) {
	return time.Parse("3PM", timeStr)
}
