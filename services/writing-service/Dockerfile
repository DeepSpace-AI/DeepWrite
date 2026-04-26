FROM postgres:18-alpine AS builder

WORKDIR /app

RUN apk add --no-cache go git

COPY go.mod ./
COPY . .

RUN go mod init github.com/deepwrite/writing-service 2>/dev/null || true
RUN go get github.com/gin-gonic/gin@latest github.com/golang-jwt/jwt@latest github.com/gorilla/websocket@latest github.com/lib/pq@latest go.mongodb.org/mongo-driver/v2@latest 2>&1 || true
RUN go build -mod=mod -o server ./cmd/server 2>&1 || go build -o server ./cmd/server

FROM postgres:18-alpine

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/server .

EXPOSE 8084

CMD ["./server"]
