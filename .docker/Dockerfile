FROM golang:1.23-alpine AS builder

WORKDIR /build
RUN apk update --no-cache && apk add --no-cache tzdata
COPY ../ .
RUN go mod tidy
RUN go build -o grpc_server cmd/server/main.go
RUN go build -o order_created_consumer cmd/transaction_inbox/order_created/main.go
RUN go build -o order_updated_consumer cmd/transaction_inbox/order_updated/main.go
RUN go build -o transaction_processed_consumer cmd/transaction_inbox/transaction_processed/main.go

FROM alpine:3.14 AS production
RUN apk update --no-cache && apk add --no-cache ca-certificates

COPY --from=builder /usr/share/zoneinfo/Europe/Moscow /usr/share/zoneinfo/Europe/Moscow
ENV TZ=Europe/Moscow
WORKDIR /build
COPY --from=builder build/grpc_server /build/grpc_server
COPY --from=builder build/order_created_consumer /build/order_created_consumer
COPY --from=builder build/order_updated_consumer /build/order_updated_consumer
COPY --from=builder build/transaction_processed_consumer /build/transaction_processed_consumer


