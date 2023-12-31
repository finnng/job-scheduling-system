# api-server/Dockerfile
FROM golang:1.21

ENV GO111MODULE=on
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o ./tmp/api-server ./api-server/app.go

EXPOSE 8081
CMD ["./tmp/api-server"]
