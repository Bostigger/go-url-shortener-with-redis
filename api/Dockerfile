# Stage 1: Build
FROM golang:alpine as builder

WORKDIR /build

COPY . .

RUN go build -o main .

# Stage 2: Run
FROM alpine

RUN adduser -S -D -H -h /app appuser

COPY --from=builder /build/main /app/
COPY --from=builder /build/.env /app/

# Switch to the appuser
USER appuser

WORKDIR /app

EXPOSE 9090

CMD ["./main"]
