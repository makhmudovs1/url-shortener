package shortener

import (
	"strconv"
	"strings"
)

var (
	str []string
	cnt int64
)

func ShortenURL(id int64) string {
	for id > 0 {
		cnt = id % 62
		id = id / 62
		str = append(str, strconv.FormatInt(int64(cnt), 10))
		cnt = id
	}
	return strings.Join(str, "")
}
