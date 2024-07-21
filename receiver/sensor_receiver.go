package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

type SensorData struct {
    SensorID    string  `json:"sensor_id"`
    Temperature float64 `json:"temperature"`
    Humidity    float64 `json:"humidity"`
}

func receiveSensorData(w http.ResponseWriter, r *http.Request) {
    var data SensorData
    err := json.NewDecoder(r.Body).Decode(&data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    fmt.Printf("Received data from %s: Temperature=%.2f, Humidity=%.2f\n", data.SensorID, data.Temperature, data.Humidity)

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, `{"status":"success"}`)
}

func main() {
    http.HandleFunc("/sensor_data", receiveSensorData)
    log.Fatal(http.ListenAndServe(":5000", nil))
}
