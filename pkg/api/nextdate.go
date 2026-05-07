package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DateFormat = "20060102"

// NextDate вычисляет следующую дату выполнения задачи.
// now - текущее время (от какой даты ищем следующую)
// dstart - дата начала/последнего выполнения задачи
// repeat - правило повторения
func NextDate(now time.Time, dstart string, repeat string) (string, error) {

	startTime, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid start date format: %w", err)
	}

	if repeat == "" {
		return "", fmt.Errorf("empty repeat rule")
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty repeat rule")
	}

	ruleType := parts[0]
	args := parts[1:]

	switch ruleType {
	case "d":
		return handleDaily(startTime, now, args)
	case "y":
		return handleYearly(startTime, now)
	case "w":
		return handleWeekly(startTime, now, args)
	case "m":
		return handleMonthly(startTime, now, args)
	default:
		return "", fmt.Errorf("unsupported repeat rule: %s", ruleType)
	}
}

// handleDaily обрабатывает правило "d <days>"
func handleDaily(start, now time.Time, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("invalid arguments for 'd' rule")
	}

	days, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid days value: %w", err)
	}

	if days <= 0 || days > 400 {
		return "", fmt.Errorf("days must be between 1 and 400")
	}

	current := start
	for {
		current = current.AddDate(0, 0, days)
		if isAfter(current, now) {
			return current.Format(DateFormat), nil
		}
	}
}

// handleYearly обрабатывает правило "y"
func handleYearly(start, now time.Time) (string, error) {
	current := start
	for {
		current = current.AddDate(1, 0, 0)
		if isAfter(current, now) {
			return current.Format(DateFormat), nil
		}
	}
}

// handleWeekly обрабатывает правило "w <days>"
// days: 1-7 (Mon-Sun)
func handleWeekly(start, now time.Time, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("invalid arguments for 'w' rule")
	}

	dayStrs := strings.Split(args[0], ",")
	validDays := make(map[int]bool)

	for _, ds := range dayStrs {
		d, err := strconv.Atoi(strings.TrimSpace(ds))
		if err != nil || d < 1 || d > 7 {
			return "", fmt.Errorf("invalid weekday value: %s", ds)
		}
		validDays[d] = true
	}

	searchStart := start
	if !isAfter(searchStart, now) {
		searchStart = now
	}

	current := searchStart.AddDate(0, 0, 1)

	limit := current.AddDate(1, 0, 0)

	for current.Before(limit) {

		goWD := int(current.Weekday())
		myWD := goWD
		if goWD == 0 {
			myWD = 7
		}

		if validDays[myWD] {
			return current.Format(DateFormat), nil
		}
		current = current.AddDate(0, 0, 1)
	}

	return "", fmt.Errorf("could not find next weekly date within a year")
}

// handleMonthly обрабатывает правило "m <days> [months]"
func handleMonthly(start, now time.Time, args []string) (string, error) {
	if len(args) < 1 || len(args) > 2 {
		return "", fmt.Errorf("invalid arguments for 'm' rule")
	}

	dayStrs := strings.Split(args[0], ",")
	validDays := make([]int, 0)
	for _, ds := range dayStrs {
		ds = strings.TrimSpace(ds)
		d, err := strconv.Atoi(ds)
		if err != nil {
			return "", fmt.Errorf("invalid day value: %s", ds)
		}
		if (d < 1 || d > 31) && d != -1 && d != -2 {
			return "", fmt.Errorf("day out of range: %d", d)
		}
		validDays = append(validDays, d)
	}

	validMonths := make(map[int]bool)
	hasMonthFilter := false
	if len(args) == 2 {
		hasMonthFilter = true
		monthStrs := strings.Split(args[1], ",")
		for _, ms := range monthStrs {
			ms = strings.TrimSpace(ms)
			m, err := strconv.Atoi(ms)
			if err != nil || m < 1 || m > 12 {
				return "", fmt.Errorf("invalid month value: %s", ms)
			}
			validMonths[m] = true
		}
	}

	searchStart := start
	if !isAfter(searchStart, now) {
		searchStart = now
	}

	current := searchStart.AddDate(0, 0, 1)

	limit := current.AddDate(2, 0, 0)

	for current.Before(limit) {
		curYear, curMonth, curDay := current.Date()
		curMonthInt := int(curMonth)

		if hasMonthFilter && !validMonths[curMonthInt] {
			current = time.Date(curYear, time.Month(curMonthInt)+1, 1, 0, 0, 0, 0, time.UTC)
			continue
		}

		if isValidDay(curDay, curYear, curMonthInt, validDays) {
			return current.Format(DateFormat), nil
		}

		current = current.AddDate(0, 0, 1)
	}

	return "", fmt.Errorf("could not find next monthly date within 2 years")
}

// isValidDay проверяет, подходит ли день под правила (включая -1, -2)
func isValidDay(day, year, month int, validDays []int) bool {
	for _, d := range validDays {
		if d > 0 {
			if d == day {
				return true
			}
		} else {
			daysInMonth := daysInMonth(year, month)
			if d == -1 && day == daysInMonth {
				return true
			}
			if d == -2 && day == daysInMonth-1 {
				return true
			}
		}
	}
	return false
}

// daysInMonth возвращает количество дней в месяце
func daysInMonth(year, month int) int {

	t := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC)
	return t.Day()
}

// isAfter проверяет, что date строго больше now.
func isAfter(date, now time.Time) bool {
	d1 := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	d2 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return d1.After(d2)
}
