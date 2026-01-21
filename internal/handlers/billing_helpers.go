package handlers

import (
	"math"
	"time"

	"codex-gateway/internal/database"
)

func roundAmount(amount float64) float64 {
	return math.Round(amount*100) / 100
}

func calculateRemainingDays(endDate time.Time, today time.Time) int {
	end := endDate.In(database.AsiaShanghai)
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, database.AsiaShanghai)
	if endDay.Before(today) {
		return 0
	}
	days := int(endDay.Sub(today).Hours() / 24)
	return days + 1
}
