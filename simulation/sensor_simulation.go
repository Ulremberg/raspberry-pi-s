package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "math/rand"
    "net/http"
    "sync"
    "time"
)

type SensorData struct {
    SensorID    string  `json:"sensor_id"`
    Temperature float64 `json:"temperature"`
    Humidity    float64 `json:"humidity"`
}

func simulateSensor(sensorID string, piIP string, duration int, wg *sync.WaitGroup) {
    defer wg.Done()

    url := fmt.Sprintf("http://%s:5000/sensor_data", piIP)

    for i := 0; i < duration; i++ {
        temperature := 20 + rand.Float64()*10
        humidity := 40 + rand.Float64()*20

        data := SensorData{
            SensorID:    sensorID,
            Temperature: temperature,
            Humidity:    humidity,
        }

        jsonData, _ := json.Marshal(data)

        http.Post(url, "application/json", bytes.NewBuffer(jsonData))
        time.Sleep(time.Second)
    }
}

func main() {
    piZeroWIP := ""  
    piZero2WIP := "" 

    numSensors := 10
    duration := 60 
    var wg sync.WaitGroup

    for _, piIP := range []string{piZeroWIP, piZero2WIP} {
        for i := 0; i < numSensors; i++ {
            sensorID := fmt.Sprintf("sensor_%d", i+1)
            wg.Add(1)
            go simulateSensor(sensorID, piIP, duration, &wg)
        }
    }

    wg.Wait()
}
