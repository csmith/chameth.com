package metrics

import (
	"regexp"
	"strconv"
)

var subscribersRe = regexp.MustCompile(`(\d+)\s+subscribers?`)

func RecordFeedRequest(feed string, userAgent string) {
	count := 1
	if matches := subscribersRe.FindStringSubmatch(userAgent); len(matches) == 2 {
		if n, err := strconv.Atoi(matches[1]); err == nil && n > 0 {
			count = n
		}
	}
	feedRequestsTotal.WithLabelValues(feed).Add(float64(count))
}
