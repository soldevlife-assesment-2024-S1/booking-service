package helpers

import "time"

func DurationCalculation(dateEnd time.Time) time.Duration {
	dateStart := time.Now()
	return dateEnd.Sub(dateStart)
}
