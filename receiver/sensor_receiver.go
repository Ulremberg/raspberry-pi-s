package main

import (
"bytes"
"crypto/aes"
"crypto/cipher"
"crypto/rand"
"encoding/json"
"fmt"
"io"
"io/ioutil"
"log"
"math"
"net/http"
"os"
"time"
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

type SoilProperties struct {
WiltingPoint float64 `json:"wilting_point"`
FieldCapacity float64 `json:"field_capacity"`
}

type SMIResult struct {
SMI float64 `json:"smi"`
SMIError float64 `json:"smi_error"`
ProcessingTime int64 `json:"processing_time_ms"`
PayloadSize int `json:"payload_size_bytes"`
}

var enableEncryption = false

func main() {
if len(os.Args) > 1 && os.Args[1] == "--encrypt" {
    enableEncryption = true
    log.Println("Processamento criptográfico habilitado")
}

http.HandleFunc("/sensor_data", receiveSensorData)

log.Println("Servidor iniciado, escutando na porta 5000")
log.Fatal(http.ListenAndServe(":5000", nil))
}

func receiveSensorData(w http.ResponseWriter, r *http.Request) {
startTime := time.Now()

body, err := ioutil.ReadAll(r.Body)
if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}

payloadSize := len(body)

var data SensorData
err = json.Unmarshal(body, &data)
if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}

if enableEncryption {
    data = processWithEncryption(data)
}

soilProps := SoilProperties{
    WiltingPoint: 0.1,
    FieldCapacity: 0.3,
}
soilMoistureError := 0.03
wiltingPointError := 0.05
fieldCapacityError := 0.12

result := calculateSMI(data.Humidity, soilProps.WiltingPoint, soilProps.FieldCapacity, 
                      soilMoistureError, wiltingPointError, fieldCapacityError)

elapsedTime := time.Since(startTime).Milliseconds()
result.ProcessingTime = elapsedTime
result.PayloadSize = payloadSize

go sendSMIToServer(data.SensorID, result)

log.Printf("Processados dados de %s em %d ms (payload: %d bytes)", 
          data.SensorID, elapsedTime, payloadSize)

w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(result)
}

func calculateSMI(soilMoisture, wiltingPoint, fieldCapacity, soilMoistureError, wiltingPointError, fieldCapacityError float64) SMIResult {
if fieldCapacity == wiltingPoint {
    return SMIResult{0, 0, 0, 0}
}

for i := 0; i < 1000; i++ {
    math.Sqrt(float64(i * i))
}

smi := (soilMoisture - wiltingPoint) / (fieldCapacity - wiltingPoint)
smiError := math.Sqrt(
    math.Pow(soilMoistureError, 2) + 
    math.Pow(wiltingPointError, 2) + 
    math.Pow(fieldCapacityError, 2))

return SMIResult{smi, smiError, 0, 0}
}

func processWithEncryption(data SensorData) SensorData {
jsonData, _ := json.Marshal(data)

key := []byte("a very very very secret key") // 32 bytes

block, _ := aes.NewCipher(key)

gcm, _ := cipher.NewGCM(block)

nonce := make([]byte, gcm.NonceSize())
if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
    panic(err)
}

ciphertext := gcm.Seal(nonce, nonce, jsonData, nil)

nonceSize := gcm.NonceSize()
if len(ciphertext) < nonceSize {
    panic("ciphertext too short")
}

nonce, ciphertext = ciphertext[:nonceSize], ciphertext[nonceSize:]
plaintext, _ := gcm.Open(nil, nonce, ciphertext, nil)

var processedData SensorData
json.Unmarshal(plaintext, &processedData)

return processedData
}

func sendSMIToServer(sensorID string, result SMIResult) {
finalServerIP := "192.168.1.200" // Substitua pelo IP real do servidor central
url := fmt.Sprintf("http://%s:8080/smi", finalServerIP)

data := map[string]interface{}{
    "sensor_id": sensorID,
    "smi": result.SMI,
    "smi_error": result.SMIError,
    "processing_time_ms": result.ProcessingTime,
    "payload_size_bytes": result.PayloadSize,
}

jsonData, err := json.Marshal(data)
if err != nil {
    log.Printf("Erro ao codificar dados SMI: %v", err)
    return
}

client := &http.Client{
    Timeout: time.Second * 10,
}

req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
if err != nil {
    log.Printf("Erro ao criar requisição: %v", err)
    return
}

req.Header.Set("Content-Type", "application/json")

resp, err := client.Do(req)
if err != nil {
    log.Printf("Erro ao enviar dados SMI para servidor: %v", err)
    return
}
defer resp.Body.Close()

log.Printf("Dados SMI enviados para servidor")
}
