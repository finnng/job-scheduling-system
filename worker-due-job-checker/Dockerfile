FROM golang:1.21

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY worker-due-job-checker/ .

RUN go build -o ./app ./app.go

CMD ["./app"]
