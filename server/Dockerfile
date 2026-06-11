FROM golang:1.23-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct
ENV GONOSUMDB=*

WORKDIR /build
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /voicechat-server .

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
RUN mkdir -p /data/certs
COPY --from=builder /voicechat-server /voicechat-server
COPY client/ /client/
EXPOSE 80 443 8080
CMD ["/voicechat-server"]
