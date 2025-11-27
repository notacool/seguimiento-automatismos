# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar binario desde build stage
COPY --from=builder /app/main .

# Exponer puerto
EXPOSE 8080

# Ejecutar aplicación
CMD ["./main"]
