package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// possible traffic light conditions
const (
	lightRed    = "red"
	lightYellow = "yellow"
	lightGreen  = "green"
)

type trafficLights struct {
	CurrentLight string `json:"currentLight"`
}

// structure for storing traffic light status
type serverState struct {
	trafficLights

	mutex sync.RWMutex
}

// traffic light instance
var lights = serverState{
	trafficLights: trafficLights{
		CurrentLight: lightRed,
	},
}

func main() {
	initServer()

	// server startup (blocking)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handlerGetLight used for obtaining the current state of the traffic light
func handlerGetLight(w http.ResponseWriter, r *http.Request) {
	lights.mutex.RLock()
	defer lights.mutex.RUnlock()

	w.Header().Add("Content-Type", "application/json")
	resp, _ := json.Marshal(lights.trafficLights)
	_, _ = w.Write(resp)
}

// handlerSetLight used for setting a new traffic light state
func handlerSetLight(w http.ResponseWriter, r *http.Request) {
	lights.mutex.Lock()
	defer lights.mutex.Unlock()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var newTrafficLights trafficLights
	err := decoder.Decode(&newTrafficLights)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validateRequest(&newTrafficLights)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lights.CurrentLight = newTrafficLights.CurrentLight
}

func validateRequest(lights *trafficLights) error {
	switch lights.CurrentLight {
	case lightRed, lightYellow, lightGreen:
		return nil
	default:
		return fmt.Errorf("incorrect current light: '%s'", lights.CurrentLight)
	}
}

func initServer() {
	http.HandleFunc("/light/get", handlerGetLight)
	http.HandleFunc("/light/set", handlerSetLight)
}
