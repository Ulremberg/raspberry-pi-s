package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
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
    WiltingPoint  float64 `json:"wilting_point"`
    FieldCapacity float64 `json:"field_capacity"`
}

type SMIResult struct {
    SMI      float64 `json:"smi"`
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
        WiltingPoint:  0.1,
        FieldCapacity: 0.3,
    }
    soilMoistureError := 0.03
    wiltingPointError := 0.05
    fieldCapacityError := 0.12
    result := calculateSMI(data.Humidity, soilProps.WiltingPoint, soilProps.FieldCapacity, soilMoistureError, wiltingPointError, fieldCapacityError)

    sendSMIToServer(data.SensorID, result)
    elapsedTime := time.Since(startTime).Milliseconds()
    log.Printf("Processed data from %s in %d ms", data.SensorID, elapsedTime)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(result)
}

func sendSMIToServer(sensorID string, result SMIResult) {
    finalServerIP := "" 
    url := fmt.Sprintf("http://%s:8080/smi", finalServerIP)
    data := map[string]interface{}{
        "sensor_id": sensorID,
        "smi":       result.SMI,
        "smi_error": result.SMIError,
    }
    jsonData, err := json.Marshal(data)
    if err != nil {
        log.Printf("Error encoding SMI data: %v", err)
        return
    }

    log.Printf("Attempting to send data to %s", url)
    log.Printf("Data: %s", string(jsonData))

    client := &http.Client{
        Timeout: time.Second * 10,
    }
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        log.Printf("Error creating request: %v", err)
        return
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error sending SMI data to server: %v", err)
        return
    }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    log.Printf("Response from server: Status: %s, Body: %s", resp.Status, string(body))

    if resp.StatusCode != http.StatusOK {
        log.Printf("Error response from server: Status: %s, Body: %s", resp.Status, string(body))
    } else {
        log.Printf("Successfully sent SMI data to server")
    }
}

func main() {
    http.HandleFunc("/sensor_data", receiveSensorData)
    log.Println("Server started, listening on :5000")
    log.Fatal(http.ListenAndServe(":5000", nil))
}