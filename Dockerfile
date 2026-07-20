# Stage 1: Build frontend
FROM node:18-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./internal/web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cloudkey .

# Stage 3: Production image
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/cloudkey .
EXPOSE 8080
ENTRYPOINT ["./cloudkey"]
