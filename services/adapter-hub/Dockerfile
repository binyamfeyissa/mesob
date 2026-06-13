FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY shared/go-kit/ shared/go-kit/
COPY shared/events/ shared/events/
COPY services/adapter-hub/ services/adapter-hub/
RUN go work init && go work use ./shared/go-kit ./shared/events ./services/adapter-hub
ENV GONOSUMDB=github.com/mesob-wallet/*
WORKDIR /build/services/adapter-hub
RUN go build -o /app ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app /app

ENTRYPOINT ["/app"]
