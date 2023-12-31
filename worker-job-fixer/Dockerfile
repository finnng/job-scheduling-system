FROM golang:1.21

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY worker-job-fixer/ .

RUN go build -o ./app ./app.go

CMD ["./app"]
