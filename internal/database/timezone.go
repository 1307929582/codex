package database

import (
	"time"
)

var AsiaShanghai *time.Location

func init() {
	var err error
	AsiaShanghai, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
}

// GetToday returns today's date in Asia/Shanghai timezone
func GetToday() time.Time {
	now := time.Now().In(AsiaShanghai)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, AsiaShanghai)
}
