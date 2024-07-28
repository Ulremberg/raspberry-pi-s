package main

import (
    "encoding/json"
    "fmt"
    "log"
    "math"
    "net/http"
    "time"
)

type SensorData struct {
    SensorID    string  `json:"sensor_id"`
    Temperature float64 `json:"temperature"`
    Humidity    float64 `json:"humidity"`
}

type SoilProperties struct {
    WiltingPoint   float64 `json:"wilting_point"`
    FieldCapacity float64 `json:"field_capacity"`
}

type SMIResult struct {
    SMI     float64 `json:"smi"`
    SMIError float64 `json:"smi_error"`
}

func calculateSMI(soilMoisture, wiltingPoint, fieldCapacity, soilMoistureError, wiltingPointError, fieldCapacityError float64) SMIResult {
    if fieldCapacity == wiltingPoint {
        return SMIResult{0, 0} 
    }

    smi := (soilMoisture - wiltingPoint) / (fieldCapacity - wiltingPoint)
   
    smiError := math.Sqrt(
        math.Pow(soilMoistureError, 2) +
        math.Pow(wiltingPointError, 2) +
        math.Pow(fieldCapacityError, 2))

    return SMIResult{smi, smiError}
}

func receiveSensorData(w http.ResponseWriter, r *http.Request) {
    startTime := time.Now()

    var data SensorData
    err := json.NewDecoder(r.Body).Decode(&data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

   
    soilProps := SoilProperties{
        WiltingPoint:   0.1, 
        FieldCapacity: 0.3, 
    }

    soilMoistureError := 0.03
    wiltingPointError := 0.05
    fieldCapacityError := 0.12

    result := calculateSMI(data.Humidity, soilProps.WiltingPoint, soilProps.FieldCapacity, soilMoistureError, wiltingPointError, fieldCapacityError)
    fmt.Printf("Received data from %s: Temperature=%.2f, Humidity=%.2f, SMI=%.2f ± %.2f\n", data.SensorID, data.Temperature, data.Humidity, result.SMI, result.SMIError)

    elapsedTime := time.Since(startTime).Milliseconds()
    log.Printf("Processed data from %s in %d ms", data.SensorID, elapsedTime)

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(result)
}

func main() {
    http.HandleFunc("/sensor_data", receiveSensorData)
    log.Fatal(http.ListenAndServe(":5000", nil))
}
