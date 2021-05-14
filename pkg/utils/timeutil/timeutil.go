package timeutil

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gidyon/micro/v2/utils/errs"
	"google.golang.org/grpc/codes"
)

// GetDateFromString converts the dateStr in format 2020-01-12 to time.Time
func GetDateFromString(dateStr string) (*time.Time, error) {
	seps := strings.Split(dateStr, "-")
	if len(seps) != 3 {
		seps = strings.Split(dateStr, "/")
		if len(seps) != 3 {
			return nil, fmt.Errorf("expected date in format 2020-00-00")
		}
	}
	sepsInt := make([]int32, 0, len(seps))
	for _, p := range seps {
		pint, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("incorrect date string")
		}
		sepsInt = append(sepsInt, int32(pint))
	}
	return ParseDayStartTime(sepsInt[0], sepsInt[1], sepsInt[2])
}

// ParseDayStartTime parses the provided year, month and day into the earliest time of the day.
func ParseDayStartTime(year, month, day int32) (*time.Time, error) {
	// 20200816204116
	// 2020y 08m 16d 20h 41m 16s
	// "2006-01-02T15:04:05Z07:00"

	monthStr := fmt.Sprint(month)
	if len(monthStr) == 1 {
		monthStr = "0" + monthStr
	}

	dayStr := fmt.Sprint(day)
	if len(dayStr) == 1 {
		dayStr = "0" + dayStr
	}

	timeRFC3339Str := fmt.Sprintf(
		"%d-%s-%sT00:00:00Z", year, monthStr, dayStr,
	)

	t, err := time.Parse(time.RFC3339, timeRFC3339Str)
	if err != nil {
		return nil, fmt.Errorf("failed to parse day time: %v", err)
	}

	return &t, nil
}

// ParseDayEndTime parses the provided year, month and day into latest time of the day
func ParseDayEndTime(year, month, day int32) (time.Time, error) {
	t1, err := ParseDayStartTime(year, month, day)
	if err != nil {
		return time.Time{}, err
	}

	t2 := t1.Add(time.Hour * 24)

	return t2, nil
}

// TimeGreaterThanCurrentMonth checks whether the month is greater than current month
func TimeGreaterThanCurrentMonth(year, month int32) error {
	// The invoice should not exceed current year and month
	currentTime := time.Now()
	currentYear := currentTime.Year()
	currentMonth := currentTime.Month()

	if int(year) > currentYear {
		return errs.WrapMessage(codes.InvalidArgument, "year greater than current year")
	}

	if int(month) > int(currentMonth) && int(year) >= currentYear {
		return errs.WrapMessage(codes.InvalidArgument, "time greater than current month")
	}

	return nil
}

// TimeGreaterThanCurrentDay checks whether the provided time details is greater than current day
func TimeGreaterThanCurrentDay(year, month, day int32) error {
	// The invoice should not exceed current year and month
	currentTime := time.Now()
	currentYear := currentTime.Year()
	currentMonth := currentTime.Month()
	currentDay := currentTime.Day()

	// Check year
	if int(year) > currentYear {
		return errs.WrapMessage(codes.InvalidArgument, "year greater than current year")
	}

	// Check month
	if int(month) > int(currentMonth) && int(year) >= currentYear {
		return errs.WrapMessage(codes.InvalidArgument, "month greater than current month")
	}

	// Check day
	if int(day) > currentDay && int(month) >= int(currentMonth) && int(year) >= currentYear {
		return errs.WrapMessage(codes.InvalidArgument, "day greater than current day")
	}

	return nil
}

// ParseDatePayload parses the date provided and fails if it's greater than current time
func ParseDatePayload(year, month, day int32) (int32, int32, int32, error) {
	var (
		YEAR, MONTH, DAY int32
		err              error
	)

	YEAR, err = ParseYear(year)
	if err != nil {
		return 0, 0, 0, err
	}
	MONTH, err = ParseMonth(month)
	if err != nil {
		return 0, 0, 0, err
	}
	DAY, err = ParseDay(day)
	if err != nil {
		return 0, 0, 0, err
	}
	err = TimeGreaterThanCurrentDay(YEAR, MONTH, DAY)
	if err != nil {
		return 0, 0, 0, err
	}
	return YEAR, MONTH, DAY, nil
}

// ParseMonthPayload parses month payload and fails if month is greater than current time
func ParseMonthPayload(year, month int32) (int32, int32, error) {
	YEAR, err := ParseYear(year)
	if err != nil {
		return 0, 0, err
	}
	MONTH, err := ParseMonth(month)
	if err != nil {
		return 0, 0, err
	}
	err = TimeGreaterThanCurrentMonth(YEAR, MONTH)
	if err != nil {
		return 0, 0, err
	}
	return YEAR, MONTH, nil
}

// ParseYear parses the year and fails it it's incorrect
func ParseYear(year int32) (int32, error) {
	switch {
	case year <= 0:
		return 0, errs.MissingField("year")
	case len(fmt.Sprint(year)) != 4:
		return 0, errs.IncorrectVal("year")
	}
	return year, nil
}

// ParseMonth parses the month and fails if it's incorrect
func ParseMonth(month int32) (int32, error) {
	switch {
	case month <= 0 || month > 12:
		return 0, errs.IncorrectVal("month")
	case len(fmt.Sprint(month)) > 2:
		return 0, errs.IncorrectVal("month")
	}
	return month, nil
}

// ParseDay parses the day and fails if it's incorrect
func ParseDay(day int32) (int32, error) {
	switch {
	case day <= 0 || day > 31:
		return 0, errs.IncorrectVal("day")
	case len(fmt.Sprint(day)) > 2:
		return 0, errs.IncorrectVal("day")
	}
	return day, nil
}
