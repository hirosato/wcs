package util

import (
	"fmt"
	"time"
)

const dateformat = "20060102"
const datetimeformat = "20060102150405"

func GetDateAndTimestamp() (string, string) {
	t := time.Now().UTC()
	millisec := int64(t.Nanosecond()) / int64(time.Millisecond)
	return t.Format(dateformat), t.Format(datetimeformat) + fmt.Sprintf("%03d", millisec)
}
