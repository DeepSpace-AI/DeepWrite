FROM postgres:18-alpine AS builder

WORKDIR /app

RUN apk add --no-cache go git

COPY go.mod go.sum ./
COPY . .
RUN go mod tidy && go build -o server ./cmd/server

FROM postgres:18-alpine

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/server .

EXPOSE 8081

CMD ["./server"]
