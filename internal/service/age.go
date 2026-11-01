package service

import "time"

// CalculateAge returns the number of full years between dob and now.
// Only the calendar date matters; time-of-day is ignored.
func CalculateAge(dob, now time.Time) int {
	age := now.Year() - dob.Year()
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}
	if age < 0 {
		return 0
	}
	return age
}
