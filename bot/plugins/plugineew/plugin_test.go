package plugineew

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jeanhua/AniaBot/common/pluginconfig"
)

func TestConfigSchema(t *testing.T) {
	p := NewPlugin()
	schema := p.ConfigSchema()

	if err := pluginconfig.RegisterStruct(schema); err != nil {
		t.Fatalf("RegisterStruct failed: %v", err)
	}

	fields := pluginconfig.Fields()
	foundEEW := false
	for _, f := range fields {
		if f.Key == "plugin.eew.enable" {
			foundEEW = true
			if f.Default != true {
				t.Errorf("expected default true for plugin.eew.enable, got %v", f.Default)
			}
		}
	}

	if !foundEEW {
		t.Errorf("plugin.eew.enable configuration field not found in registered fields")
	}
}

func TestGetWSURL(t *testing.T) {
	p := NewPlugin()

	tests := []struct {
		source   string
		expected string
	}{
		{"sc_eew", "wss://ws-api.wolfx.jp/sc_eew"},
		{"cenc_eew", "wss://ws-api.wolfx.jp/cenc_eew"},
		{"jma_eew", "wss://ws-api.wolfx.jp/jma_eew"},
		{"fj_eew", "wss://ws-api.wolfx.jp/fj_eew"},
		{"cq_eew", "wss://ws-api.wolfx.jp/cq_eew"},
		{"all_eew", "wss://ws-api.wolfx.jp/all_eew"},
		{"unknown", "wss://ws-api.wolfx.jp/sc_eew"},
	}

	for _, tc := range tests {
		got := p.getWSURL(tc.source)
		if got != tc.expected {
			t.Errorf("getWSURL(%s) = %s, expected %s", tc.source, got, tc.expected)
		}
	}
}

func TestEEWEventParsing(t *testing.T) {
	rawJSON := `{
		"type": "sc_eew",
		"ID": 12345,
		"EventID": "SC2026072801",
		"ReportNum": 1,
		"ReportTime": "2026-07-28 11:30:00",
		"OriginTime": "2026-07-28 11:29:55",
		"HypoCenter": "四川宜宾市高县",
		"Latitude": 28.51,
		"Longitude": 104.67,
		"Magunitude": 3.8,
		"MaxIntensity": 5,
		"isFinal": false
	}`

	var event EEWEvent
	if err := json.Unmarshal([]byte(rawJSON), &event); err != nil {
		t.Fatalf("Unmarshal EEWEvent error: %v", err)
	}

	if event.EventID != "SC2026072801" {
		t.Errorf("expected EventID SC2026072801, got %s", event.EventID)
	}

	if event.GetMagnitude() != 3.8 {
		t.Errorf("expected magnitude 3.8, got %f", event.GetMagnitude())
	}

	if event.GetLocation() != "四川宜宾市高县" {
		t.Errorf("expected location 四川宜宾市高县, got %s", event.GetLocation())
	}

	if event.GetMaxIntensityStr() != "5" {
		t.Errorf("expected max intensity 5, got %s", event.GetMaxIntensityStr())
	}
}

func TestWeatherRankParsing(t *testing.T) {
	rawJSON := `{
		"202607281100": {
			"tempRank": [
				{"province": "四川", "city": "成都", "value": "35.2 ℃"}
			],
			"rainRank": [],
			"windSRank": []
		}
	}`

	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(rawJSON), &rawMap); err != nil {
		t.Fatalf("Unmarshal weather rank rawMap error: %v", err)
	}

	if _, ok := rawMap["202607281100"]; !ok {
		t.Fatalf("expected key 202607281100 in map")
	}

	var hourRank WeatherHourRank
	if err := json.Unmarshal(rawMap["202607281100"], &hourRank); err != nil {
		t.Fatalf("Unmarshal WeatherHourRank error: %v", err)
	}

	if len(hourRank.TempRank) != 1 || hourRank.TempRank[0].City != "成都" {
		t.Errorf("expected tempRank item for 成都, got %v", hourRank.TempRank)
	}
}

func TestWSDial(t *testing.T) {
	p := NewPlugin()
	urlStr := p.getWSURL("cenc_eew")
	t.Logf("Testing dial to %s", urlStr)

	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 15 * time.Second,
	}

	header := make(http.Header)
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	header.Set("Origin", "https://wolfx.jp")
	header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	conn, resp, err := dialer.Dial(urlStr, header)
	if err != nil {
		if resp != nil {
			t.Logf("Dial returned status %s (%d): %v", resp.Status, resp.StatusCode, err)
		} else {
			t.Logf("Dial failed: %v", err)
		}
		return
	}
	defer conn.Close()
	t.Logf("Successfully connected to %s", urlStr)
}
