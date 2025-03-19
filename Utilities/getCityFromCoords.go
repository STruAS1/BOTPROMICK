package Utilities

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func GetCityFromCoords(lat, lon float64) (string, error) {
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=json&lat=%f&lon=%f", lat, lon)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Address struct {
			City string `json:"city"`
		} `json:"address"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	if result.Address.City != "" {
		return result.Address.City, nil
	}
	return "Неизвестный город", nil
}
