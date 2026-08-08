package request

import (
	"strconv"
	"strings"
)

func PathID(path, prefix string) (int, bool) {
	value := strings.TrimPrefix(path, prefix)
	id, err := strconv.Atoi(value)
	return id, value != "" && err == nil
}
