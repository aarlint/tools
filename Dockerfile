FROM golang:1.24-alpine AS go-build
WORKDIR /build
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/main.go .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o tools-server .

FROM node:22-alpine AS frontend-build
WORKDIR /build
COPY package.json package-lock.json* ./
RUN npm ci
COPY . .
RUN npm run build

FROM golang:1.24-alpine AS prep
RUN mkdir /data && chown 65532:65532 /data

FROM cgr.dev/chainguard/static:latest
COPY --from=prep /data /data
COPY --from=go-build /build/tools-server /tools-server
COPY --from=frontend-build /build/dist /dist
ENV DIST_DIR=/dist
ENV DATA_DIR=/data
ENV ADDR=:8080
ENTRYPOINT ["/tools-server"]
