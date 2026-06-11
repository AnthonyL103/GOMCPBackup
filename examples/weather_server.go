// testing/test_server.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	// Route 1: get_weather handler
	http.HandleFunc("/execute/get_weather", handleGetWeather)
	
	// Route 2: get_forecast handler
	http.HandleFunc("/execute/get_forecast", handleGetForecast)
	
	// Hardcoded port matching YAML config
	log.Println("Starting test weather server on port 8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleGetWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	city, _ := params["city"].(string)
	units, _ := params["units"].(string)
	
	if units == "" {
		units = "fahrenheit"
	}
	
	temp := 65
	if units == "celsius" {
		temp = 18
	}
	
	result := map[string]interface{}{
		"city":        city,
		"temperature": temp,
		"condition":   "sunny",
		"units":       units,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleGetForecast(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	city, _ := params["city"].(string)
	days, _ := params["days"].(float64)
	
	if days == 0 {
		days = 1
	}
	
	forecast := []map[string]interface{}{}
	for i := 0; i < int(days); i++ {
		forecast = append(forecast, map[string]interface{}{
			"day":       i + 1,
			"temp":      65 + i,
			"condition": "sunny",
		})
	}
	
	result := map[string]interface{}{
		"city":     city,
		"forecast": forecast,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}