package openweathermap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OpenWeatherMap struct {
	ApiKey string
}

type owmResponse struct {
	Message string `json:"message"`
	Sys     struct {
		Sunset int64 `json:"sunset"`
	} `json:"sys"`
}

func New(apiKey string) *OpenWeatherMap {
	return &OpenWeatherMap{
		ApiKey: apiKey,
	}
}

func (owm *OpenWeatherMap) SunsetTime(city string) (time.Time, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s", city, owm.ApiKey)
	resp, err := http.Get(url)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	weather := new(owmResponse)
	if err := json.NewDecoder(resp.Body).Decode(weather); err != nil {
		return time.Time{}, err
	}

	if weather.Message != "" {
		return time.Time{}, fmt.Errorf("error from openweathermap: %v", err)
	}

	return time.Unix(weather.Sys.Sunset, 0), nil
}
