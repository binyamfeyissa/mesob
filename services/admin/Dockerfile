FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY shared/go-kit/ shared/go-kit/
COPY shared/events/ shared/events/
COPY services/admin/ services/admin/
RUN go work init && go work use ./shared/go-kit ./shared/events ./services/admin
ENV GONOSUMDB=github.com/mesob-wallet/*
WORKDIR /build/services/admin
RUN go build -o /app ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app /app
EXPOSE 8009 9109
ENTRYPOINT ["/app"]
