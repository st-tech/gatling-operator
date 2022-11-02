package utils

import (
	"hash/fnv"
	"strings"
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
func Add_labels_pods(pod_type string, pod_obectmeta map[string]string) map[string]string {
	if pod_obectmeta != nil {
		if strings.Contains(pod_type, "runner") == true {
			pod_obectmeta["type"] = "runner"
		} else if strings.Contains(pod_type, "reporter") == true {
			pod_obectmeta["type"] = "reporter"
		}
		return pod_obectmeta
	} else {
		pod_obectmeta := map[string]string{}
		if strings.Contains(pod_type, "runner") == true {
			pod_obectmeta["type"] = "runner"
		} else if strings.Contains(pod_type, "reporter") == true {
			pod_obectmeta["type"] = "reporter"
		}
		return pod_obectmeta
	}
}
