package utils

import (
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
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
	if dataMap == nil {
		dataMap = map[string]string{}
	}
	dataMap[key] = value
	return dataMap

}

func GetNumEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	valueNum, err := strconv.Atoi(value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ENV '%s' is not valid\n", key)
		return defaultValue
	}
	return valueNum
}
