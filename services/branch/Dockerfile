FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY shared/go-kit/ shared/go-kit/
COPY shared/events/ shared/events/
COPY services/branch/ services/branch/
RUN go work init && go work use ./shared/go-kit ./shared/events ./services/branch
ENV GONOSUMDB=github.com/mesob-wallet/*
WORKDIR /build/services/branch
RUN go build -o /app ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app /app
EXPOSE 8008 9108
ENTRYPOINT ["/app"]
