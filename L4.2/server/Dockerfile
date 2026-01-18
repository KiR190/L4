FROM golang:1.25-alpine

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/grep-server ./server/main.go

EXPOSE 50053
ENTRYPOINT ["/app/grep-server"]
