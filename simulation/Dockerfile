FROM golang:latest

WORKDIR /app

COPY . .

RUN go build -o sensor_simulation sensor_simulation.go

CMD ["./sensor_simulation"]
