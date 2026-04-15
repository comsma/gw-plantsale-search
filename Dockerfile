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
COPY --from=css-builder /app/ui/static/css ./ui/static/css
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /plantsale-search ./cmd/web

# Stage 3: Runtime — assets are embedded via embed.FS, binary is self-contained
FROM scratch
COPY --from=go-builder /plantsale-search /plantsale-search

EXPOSE 8080
ENTRYPOINT ["/plantsale-search"]
