FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
RUN go get && go build -o discord_vad main.go

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/discord_vad .
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
EXPOSE 8080
CMD ["./discord_vad"]
