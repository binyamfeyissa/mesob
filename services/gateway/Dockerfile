FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY shared/go-kit/ shared/go-kit/
COPY shared/events/ shared/events/
COPY services/gateway/ services/gateway/
RUN go work init && go work use ./shared/go-kit ./shared/events ./services/gateway
ENV GONOSUMDB=github.com/mesob-wallet/*
WORKDIR /build/services/gateway
RUN go build -o /app ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app /app
EXPOSE 8000
ENTRYPOINT ["/app"]
