package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Order struct {
	OrderID   string `json:"orderId"`
	DeviceID  string `json:"deviceId"`
	PackageID string `json:"packageId"`
	Money     string `json:"money"`
	Status    string `json:"status"`
	CardKey   string `json:"cardKey,omitempty"`
	Created   int64  `json:"created"`
}

type Card struct {
	CardKey    string `json:"cardKey"`
	OrderID    string `json:"orderId"`
	BindDevice string `json:"bindDevice"`
	ChangeLeft int    `json:"changeLeft"`
	Created    int64  `json:"created"`
}

func kvCmd(args ...any) (string, error) {
	base := strings.TrimRight(os.Getenv("KV_REST_API_URL"), "/")
	token := os.Getenv("KV_REST_API_TOKEN")
	if base == "" || token == "" {
		return "", fmt.Errorf("KV_REST_API_URL / KV_REST_API_TOKEN 未配置")
	}
	body, err := json.Marshal(args)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, base, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("kv http %d: %s", resp.StatusCode, string(raw))
	}
	var wrap struct {
		Result any `json:"result"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return "", err
	}
	if wrap.Result == nil {
		return "", nil
	}
	switch v := wrap.Result.(type) {
	case string:
		return v, nil
	default:
		b, _ := json.Marshal(v)
		return string(b), nil
	}
}

func SetJSON(key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = kvCmd("SET", key, string(b))
	return err
}

func GetJSON(key string, dest any) error {
	s, err := kvCmd("GET", key)
	if err != nil {
		return err
	}
	if s == "" || s == "null" {
		return fmt.Errorf("not found")
	}
	return json.Unmarshal([]byte(s), dest)
}
