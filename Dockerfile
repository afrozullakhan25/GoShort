FROM golang:alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o goshort_binary main.go

# --- Stage 2: Runner ---
FROM alpine:3.22.2

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/goshort_binary .

EXPOSE 8080

CMD ["sh", "-c", "./goshort_binary"]
