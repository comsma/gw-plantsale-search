# Stage 1: Build CSS
FROM node:22-alpine AS css-builder
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install
COPY ui/static/styles ./ui/static/styles
COPY ui/views ./ui/views
RUN npm run css

# Stage 2: Build Go binary
FROM golang:1.26-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /portfolio ./cmd/web

# Stage 3: Runtime
FROM scratch
WORKDIR /app
COPY --from=go-builder /portfolio /portfolio
COPY --from=go-builder /app/ui/static /app/ui/static
COPY --from=go-builder /app/ui/views /app/ui/views
COPY --from=css-builder /app/ui/static/css /app/ui/static/css

EXPOSE 8080
ENTRYPOINT ["/search"]
