package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

type SMIData struct {
    SensorID string  `json:"sensor_id"`
    SMI      float64 `json:"smi"`
    SMIError float64 `json:"smi_error"`
}

func main() {
    http.HandleFunc("/smi", receiveSMIData)  
    log.Println("Server started, listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func receiveSMIData(w http.ResponseWriter, r *http.Request) {
    var smiData SMIData
    err := json.NewDecoder(r.Body).Decode(&smiData)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    log.Printf("Received SMI data: Sensor ID=%s, SMI=%.2f, SMI Error=%.2f", smiData.SensorID, smiData.SMI, smiData.SMIError)
    fmt.Fprintf(w, "SMI data received successfully")
}
