package handlers

import (
	"errors"
	"strconv"
	"strings"
)

func redisInfoMap(raw string) map[string]string {
	values := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if ok {
			values[key] = value
		}
	}
	return values
}

func monitorCacheKeyQuery(rawLimit string, rawPattern string) (int64, string, error) {
	limit, _ := strconv.ParseInt(rawLimit, 10, 64)
	pattern := strings.TrimSpace(rawPattern)
	if pattern == "" {
		pattern = "*"
	}
	if len(pattern) > 128 {
		return 0, "", errors.New("pattern 过长")
	}
	if strings.ContainsAny(pattern, "\n\r\x00") {
		return 0, "", errors.New("pattern 非法")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return limit, pattern, nil
}
