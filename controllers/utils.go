package controllers

import (
	"encoding/json"
	"hash/fnv"
	"net/http"
	"net/url"
	"time"
)

func hash(s string) uint32 {
	h := fnv.New32a()
	if _, err := h.Write([]byte(s)); err != nil {
		return 0
	}
	return h.Sum32()
}

func getEpocTime() int32 {
	return int32(time.Now().Unix())
}

type payload struct {
	Text string `json:"text"`
}

func getMapValue(key string, dataMap map[string][]byte) (string, bool) {
	if v, exists := dataMap[key]; exists {
		return string(v), true
	}
	return "", false
}

func slackNotify(webhookURL string, payloadText string) error {
	p := payload{
		Text: payloadText,
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	resp, err := http.PostForm(webhookURL,
		url.Values{"payload": {string(s)}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
