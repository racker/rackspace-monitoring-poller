package utils

import "time"

type NowTimestampMillisFunc func() int64

var NowTimestampMillis NowTimestampMillisFunc = func() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
