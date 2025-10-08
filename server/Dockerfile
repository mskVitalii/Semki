# build stage
FROM golang:1.25.1-alpine3.22 AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main ./cmd/main.go

# run stage
FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/app.yml .

EXPOSE 8080
CMD ["/app/main"]

#HEALTHCHECK --interval=3s --timeout=3s --start-period=1s --retries=3 \
#  CMD wget -q --spider http://localhost:8080/service/healthcheck || exit 1