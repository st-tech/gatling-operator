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

// Add key-value to map
func AddMapValue(key string, value string, dataMap map[string]string, overwrite bool) map[string]string {
	if _, ok := dataMap[key]; ok && !overwrite {
		return dataMap
	}
	dataMap[key] = value
	return dataMap
}
