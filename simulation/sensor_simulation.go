package main

import (
"bytes"
"encoding/json"
"flag"
"fmt"
"log"
"math/rand"
"net/http"
"sync"
"time"
"crypto/aes"
"crypto/cipher"
"crypto/rand"
"io"
)

type SensorData struct {
SensorID    string             `json:"sensor_id"`
Temperature float64            `json:"temperature"`
Humidity    float64            `json:"humidity"`
Timestamp   int64              `json:"timestamp"`
Location    [2]float64         `json:"location"`     
Metadata    map[string]string  `json:"metadata"`      
Readings    []float64          `json:"readings"`      
}

func main() {
devicePtr := flag.String("device", "zero_w", "Dispositivo alvo (zero_w ou zero_2w)")
scenarioPtr := flag.String("scenario", "default", "Nome do cenário de teste")
sensorsPtr := flag.Int("sensors", 100, "Número de sensores")
payloadPtr := flag.String("payload", "small", "Tamanho do payload (small, medium, large, xlarge)")
frequencyPtr := flag.Int("frequency", 500, "Frequência de envio em ms")
durationPtr := flag.Int("duration", 300, "Duração do teste em segundos")
flag.Parse()

var piIP string
if *devicePtr == "zero_w" {
    piIP = "192." 
} else {
    piIP = "192." 
}

log.Printf("Iniciando simulação para %s: %d sensores, payload %s, frequência %dms, duração %ds",
    *devicePtr, *sensorsPtr, *payloadPtr, *frequencyPtr, *durationPtr)

dataCh := make(chan SensorData, 1000)
var wg sync.WaitGroup
sendersWg := sync.WaitGroup{}

clients := make([]*http.Client, 10)
for i := range clients {
    clients[i] = &http.Client{
        Timeout: time.Second * 30,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 100,
            IdleConnTimeout:     90 * time.Second,
        },
    }
}

for i := range clients {
    sendersWg.Add(1)
    go sendSensorData(piIP, dataCh, &sendersWg, clients[i])
}

for i := 0; i < *sensorsPtr; i++ {
    sensorID := fmt.Sprintf("sensor_%d", i+1)
    wg.Add(1)
    go simulateSensorWithConfig(sensorID, dataCh, &wg, *durationPtr, *payloadPtr, time.Duration(*frequencyPtr))
}

wg.Wait()
close(dataCh)
sendersWg.Wait()

log.Printf("Simulação concluída para %s", *devicePtr)
}

func simulateSensorWithConfig(sensorID string, dataCh chan SensorData, wg *sync.WaitGroup, duration int, payloadSize string, frequency time.Duration) {
defer wg.Done()

startTime := time.Now()
for time.Since(startTime).Seconds() < float64(duration) {
    temperature := 20 + rand.Float64()*10
    humidity := 40 + rand.Float64()*20
    
    data := SensorData{
        SensorID:    sensorID,
        Temperature: temperature,
        Humidity:    humidity,
        Timestamp:   time.Now().Unix(),
        Location:    [2]float64{rand.Float64()*180 - 90, rand.Float64()*360 - 180},
    }
    
   
    switch payloadSize {
    case "small":        
    case "medium": 
        data.Metadata = generateMetadata(5)
        data.Readings = generateHistoricalReadings(10)
    case "large": 
        data.Metadata = generateMetadata(20)
        data.Readings = generateHistoricalReadings(100)
    case "xlarge": 
        data.Metadata = generateMetadata(50)
        data.Readings = generateHistoricalReadings(1000)
    }
    
    dataCh <- data
    time.Sleep(frequency * time.Millisecond)
}
}

func generateMetadata(count int) map[string]string {
metadata := make(map[string]string)
for i := 0; i < count; i++ {
    key := fmt.Sprintf("meta_%d", i)
    value := fmt.Sprintf("value_%d_%s", i, randomString(20))
    metadata[key] = value
}
return metadata
}

func generateHistoricalReadings(count int) []float64 {
readings := make([]float64, count)
for i := 0; i < count; i++ {
    readings[i] = rand.Float64() * 100
}
return readings
}

func randomString(length int) string {
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
b := make([]byte, length)
for i := range b {
    b[i] = charset[rand.Intn(len(charset))]
}
return string(b)
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
        log.Printf("Erro ao enviar dados de %s: %v", data.SensorID, err)
    } else {
        resp.Body.Close()
        log.Printf("Dados enviados de %s para %s em %d ms", data.SensorID, piIP, elapsedTime)
    }
}
}
