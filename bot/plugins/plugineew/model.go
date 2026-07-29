package plugineew

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// EEWEvent WebSocket / HTTP JSON 推送体模型
type EEWEvent struct {
	Type         string   `json:"type"`
	Title        string   `json:"Title"`
	EventID      string   `json:"EventID"`
	ReportNum    int      `json:"ReportNum"`
	ReportTime   string   `json:"ReportTime"`
	OriginTime   string   `json:"OriginTime"`
	HypoCenter   string   `json:"HypoCenter"`
	Hypocenter   string   `json:"Hypocenter"`
	Latitude     float64  `json:"Latitude"`
	Longitude    float64  `json:"Longitude"`
	Magnitude    float64  `json:"Magnitude"`
	Magunitude   float64  `json:"Magunitude"`
	Depth        *float64 `json:"Depth"`
	MaxIntensity any      `json:"MaxIntensity"`
	IsFinal      bool     `json:"isFinal"`
	IsCancel     bool     `json:"isCancel"`
	OriginalText string   `json:"OriginalText"`
	Ver          int      `json:"ver"`
	ID           any      `json:"id"`
	Timestamp    any      `json:"timestamp"`
}

func (e *EEWEvent) GetMagnitude() float64 {
	if e.Magnitude > 0 {
		return e.Magnitude
	}
	return e.Magunitude
}

func (e *EEWEvent) GetLocation() string {
	if e.HypoCenter != "" {
		return e.HypoCenter
	}
	if e.Hypocenter != "" {
		return e.Hypocenter
	}
	if e.Title != "" {
		return e.Title
	}
	return "未知震源地"
}

func (e *EEWEvent) GetDepthStr() string {
	if e.Depth == nil {
		return "未知"
	}
	return fmt.Sprintf("%.1f km", *e.Depth)
}

func (e *EEWEvent) GetMaxIntensityStr() string {
	if e.MaxIntensity == nil {
		return "未知"
	}
	switch v := e.MaxIntensity.(type) {
	case string:
		return v
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return fmt.Sprintf("%.1f", v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (e *EEWEvent) GetIntensityNum() int {
	if e.MaxIntensity == nil {
		return 0
	}
	switch v := e.MaxIntensity.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(v)
		return n
	default:
		return 0
	}
}

// GetEventTime 超强全兼容时间解析：支持多种时间格式及 EventID (如 202607130727.0001_1) 时间前缀解析
func (e *EEWEvent) GetEventTime() time.Time {
	// 1. 优先尝试从 ReportTime 或 OriginTime 解析
	for _, timeStr := range []string{e.ReportTime, e.OriginTime} {
		timeStr = strings.TrimSpace(timeStr)
		if timeStr == "" {
			continue
		}
		timeStr = strings.ReplaceAll(timeStr, "/", "-")

		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02 15:04:05.000",
			"2006-01-02 15:04",
			"2006-01-02T15:04:05Z07:00",
		}
		for _, fmtStr := range formats {
			if t, err := time.ParseInLocation(fmtStr, timeStr, time.Local); err == nil {
				return t
			}
		}

		if sec, err := strconv.ParseInt(timeStr, 10, 64); err == nil && sec > 1000000000 {
			return time.Unix(sec, 0)
		}
	}

	// 2. 从 EventID 提取年月日时分 (如 202607130727.0001_1 提取前12位数字 202607130727)
	eventID := strings.TrimSpace(e.EventID)
	if len(eventID) >= 12 {
		var dateDigits strings.Builder
		for _, r := range eventID {
			if r >= '0' && r <= '9' {
				dateDigits.WriteRune(r)
				if dateDigits.Len() == 12 {
					break
				}
			} else {
				break
			}
		}
		if dateDigits.Len() == 12 {
			if t, err := time.ParseInLocation("200601021504", dateDigits.String(), time.Local); err == nil {
				return t
			}
		}
	}

	return time.Time{}
}

// CENCEQItem 中国地震台网 历史速报项
type CENCEQItem struct {
	Type       string `json:"type"`
	EventID    string `json:"EventID"`
	Time       string `json:"time"`
	ReportTime string `json:"ReportTime"`
	Location   string `json:"location"`
	PlaceName  string `json:"placeName"`
	Magnitude  string `json:"magnitude"`
	Depth      string `json:"depth"`
	Latitude   string `json:"latitude"`
	Longitude  string `json:"longitude"`
	Intensity  string `json:"intensity"`
}

// WeatherRankItem 气象排行元素
type WeatherRankItem struct {
	Province string `json:"province"`
	City     string `json:"city"`
	Value    string `json:"value"`
}

// WeatherHourRank 每小时三大气象排行
type WeatherHourRank struct {
	TempRank  []WeatherRankItem `json:"tempRank"`
	RainRank  []WeatherRankItem `json:"rainRank"`
	WindSRank []WeatherRankItem `json:"windSRank"`
}
