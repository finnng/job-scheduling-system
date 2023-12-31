FROM golang:1.21

ENV GO111MODULE=on
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o ./tmp/worker-due-job-checker ./worker-due-job-checker/app.go
CMD ["./tmp/worker-due-job-checker"]
