FROM golang:1.16-alpine

WORKDIR /app

COPY . .

RUN go build -o server server.go

EXPOSE 8080

CMD ["./server"]
