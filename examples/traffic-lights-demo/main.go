package main

import (
	"encoding/json"
	"fmt"
	"io"
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

// structure for storing traffic light status
type trafficLights struct {
	CurrentLight string       `json:"currentLight"`
	mutex        sync.RWMutex `json:"-"`
}

// traffic light instance
var lights = trafficLights{
	CurrentLight: lightRed,
}

func main() {
	initServer()

	// server startup (blocking)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initServer() {
	// method for obtaining the current state of the traffic light
	http.HandleFunc("/light/get", func(w http.ResponseWriter, r *http.Request) {
		lights.mutex.RLock()
		defer lights.mutex.RUnlock()

		resp, err := json.Marshal(lights)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(resp)
	})

	// method for setting a new traffic light state
	http.HandleFunc("/light/set", func(w http.ResponseWriter, r *http.Request) {
		lights.mutex.Lock()
		defer lights.mutex.Unlock()

		request, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		var newTrafficLights trafficLights
		if err := json.Unmarshal(request, &newTrafficLights); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := validateRequest(&newTrafficLights); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		lights.CurrentLight = newTrafficLights.CurrentLight
	})
}

func validateRequest(lights *trafficLights) error {
	if lights.CurrentLight != lightRed &&
		lights.CurrentLight != lightYellow &&
		lights.CurrentLight != lightGreen {
		return fmt.Errorf("incorrect current light: %s", lights.CurrentLight)
	}
	return nil
}
