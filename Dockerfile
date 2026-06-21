FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server

FROM alpine
RUN mkdir /data
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 50051 8000
CMD ["./server"]