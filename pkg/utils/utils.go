package utils

import (
	"hash/fnv"
	"time"
)

func Hash(s string) uint32 {
	h := fnv.New32a()
	if _, err := h.Write([]byte(s)); err != nil {
		return 0
	}
	return h.Sum32()
}

func GetEpocTime() int32 {
	return int32(time.Now().Unix())
}

func GetMapValue(key string, dataMap map[string][]byte) (string, bool) {
	if v, exists := dataMap[key]; exists {
		return string(v), true
	}
	return "", false
}

// Determine whether the label is attached to the runner or the reporter.
func AddMapValue(key string, value string, dataMap map[string]string, overwrite bool) map[string]string {
	_, ok := dataMap[key]
	if overwrite {
		if ok {
			dataMap[key] = value
			return dataMap
		} else {
			dataMap = map[string]string{}
			dataMap[key] = value
			return dataMap
		}
	} else {
		if ok {
			return dataMap
		} else {
			dataMap = map[string]string{}
			dataMap[key] = value
			return dataMap
		}
	}
}
