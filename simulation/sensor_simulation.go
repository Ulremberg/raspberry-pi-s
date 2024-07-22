package main

import (
	"bytes"
    "encoding/json"
    "fmt"
    "log"
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

        startTime := time.Now()
        resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
        elapsedTime := time.Since(startTime).Milliseconds()

        if err != nil {
            log.Printf("Error sending data from %s: %v", sensorID, err)
        } else {
            resp.Body.Close()
            log.Printf("Sent data from %s to %s in %d ms", sensorID, piIP, elapsedTime)
        }

        time.Sleep(time.Second)
    }
}

func main() {
	
    piZeroWIP := ""  
    //piZero2WIP := "" 

    numSensors := 10
    duration := 60 
    var wg sync.WaitGroup

    for _, piIP := range []string{piZeroWIP} {
        for i := 0; i < numSensors; i++ {
            sensorID := fmt.Sprintf("sensor_%d", i+1)
            wg.Add(1)
            go simulateSensor(sensorID, piIP, duration, &wg)
        }
    }

    wg.Wait()
}
