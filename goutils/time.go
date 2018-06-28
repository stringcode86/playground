package goutils

import (
	"time"
)

// UnixMilli creates time.Time from milliseconds time stamp ie 1523098153599
func UnixMilli(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

// Time alias for `time.Time`
type Time time.Time

// UnixMilli number of miliseconds since 1970
func (t Time) UnixMilli() int64 {
	return (time.Time)(t).UnixNano() / int64(time.Millisecond)
}

//
func (t Time) UnixFloatNano() float64 {
	return float64(time.Time(t).UnixNano()) / float64(time.Second)
}

type Duration time.Duration

func (d Duration) UnixFloatNano() float64 {
	return float64(time.Duration(d).Nanoseconds()) / float64(time.Second)
}
