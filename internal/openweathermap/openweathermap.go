package openweathermap

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type OpenWeatherMap struct {
	apiKey string
}

type owmResponse struct {
	Message string `json:"message"`
	Sys     struct {
		Sunset int64 `json:"sunset"`
	} `json:"sys"`
}

func New(apiKey string) OpenWeatherMap {
	return OpenWeatherMap{apiKey}
}

func (owm OpenWeatherMap) SunsetTime(city string) (time.Time, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s", city, owm.apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	var sunset owmResponse
	if err := json.NewDecoder(resp.Body).Decode(&sunset); err != nil {
		return time.Time{}, err
	}

	if sunset.Message != "" {
		return time.Time{}, errors.New(sunset.Message)
	}

	return time.Unix(sunset.Sys.Sunset, 0), nil
}
