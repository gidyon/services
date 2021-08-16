package timeutil

import (
	"time"

	"github.com/gidyon/micro/v2/utils/errs"
	"google.golang.org/grpc/codes"
)

func FetchMinMaxDates(date1, date2 string) ([]*time.Time, error) {
	t1, err := GetDateFromString(date1)
	if err != nil {
		return nil, err
	}
	t2, err := GetDateFromString(date2)
	if err != nil {
		return nil, err
	}
	var max, min *time.Time
	if t1.Unix() > t2.Unix() {
		max = t1
		min = t2
	} else {
		max = t2
		min = t1
	}

	return []*time.Time{min, max}, nil
}

func GetDateRanges(date1, date2 string) ([]string, error) {
	t1, err := GetDateFromString(date1)
	if err != nil {
		return nil, err
	}
	t2, err := GetDateFromString(date2)
	if err != nil {
		return nil, err
	}

	var max, min *time.Time
	if t1.Unix() > t2.Unix() {
		max = t1
		min = t2
	} else {
		max = t2
		min = t1
	}

	// High future dates not allowed
	if max.Unix() > time.Now().Add(24*time.Hour).Unix() {
		return nil, errs.WrapMessagef(codes.InvalidArgument, "time %s is greater than todays range", max.String()[:10])
	}

	dates := []string{}

	for t := *min; t.Unix() <= max.Unix(); t = t.Add(24 * time.Hour) {
		dates = append(dates, t.String()[:10])
	}

	return dates, nil
}
