FROM golang:1.21 as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY main.go .
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o todoist-assistant .

FROM alpine:latest

COPY --from=builder /app/todoist-assistant .

ENTRYPOINT ["./todoist-assistant"]