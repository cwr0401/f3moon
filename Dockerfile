# ---- Stage 1: Build Vue frontend ----
FROM node:24-alpine AS web-builder

WORKDIR /build/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# ---- Stage 2: Build Go server ----
FROM golang:1.26-alpine AS go-builder

WORKDIR /build
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY . .
COPY --from=web-builder /build/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /f3moon_server ./cmd/server

# ---- Stage 3: Runtime ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=go-builder /f3moon_server .
COPY web/templates ./web/templates
COPY --from=web-builder /build/web/dist ./web/dist

ENV SERVER_PORT=8080 \
    SERVER_HOST=0.0.0.0 \
    GIN_MODE=release

EXPOSE 8080
ENTRYPOINT ["./f3moon_server"]
