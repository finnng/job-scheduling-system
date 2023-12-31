FROM golang:1.21

ENV GO111MODULE=on
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o ./tmp/worker-job-fixer ./worker-job-fixer/app.go

CMD ["./tmp/worker-job-fixer"]
