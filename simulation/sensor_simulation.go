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

func simulateSensor(sensorID, piIP string, dataCh chan SensorData, wg *sync.WaitGroup, duration int) {
    defer wg.Done()

    startTime := time.Now()
    for time.Since(startTime).Seconds() < float64(duration) {
        temperature := 20 + rand.Float64()*10
        humidity := 40 + rand.Float64()*20

        data := SensorData{
            SensorID:    sensorID,
            Temperature: temperature,
            Humidity:    humidity,
        }

        dataCh <- data
    }
}

func sendSensorData(piIP string, dataCh <-chan SensorData, wg *sync.WaitGroup, client *http.Client) {
    defer wg.Done()

    url := fmt.Sprintf("http://%s:5000/sensor_data", piIP)

    for data := range dataCh {
        jsonData, _ := json.Marshal(data)

        startTime := time.Now()
        resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
        elapsedTime := time.Since(startTime).Milliseconds()

        if err != nil {
            log.Printf("Error sending data from %s: %v", data.SensorID, err)
        } else {
            resp.Body.Close()
            log.Printf("Sent data from %s to %s in %d ms", data.SensorID, piIP, elapsedTime)
        }
    }
}

func main() {
    piZeroWIP := ""
    // piZero2WIP := ""

    numSensors := 100
    duration := 120 
    var wg sync.WaitGroup

    dataCh := make(chan SensorData, 1000)
    sendersWg := sync.WaitGroup{}
 
    clients := make([]*http.Client, 10)
    for i := range clients {
        clients[i] = &http.Client{}
    }
    
    for _, piIP := range []string{piZeroWIP} {
        for i := range clients {
            sendersWg.Add(1)
            go sendSensorData(piIP, dataCh, &sendersWg, clients[i])
        }
    }
   
    for i := 0; i < numSensors; i++ {
        sensorID := fmt.Sprintf("sensor_%d", i+1)
        wg.Add(1)
        go simulateSensor(sensorID, piZeroWIP, dataCh, &wg, duration)
    }

    wg.Wait()
    close(dataCh)
    sendersWg.Wait()
}
