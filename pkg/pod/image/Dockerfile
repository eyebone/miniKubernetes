### build stage ###
FROM golang:1.14-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server

### run stage ###
FROM alpine:3.12
WORKDIR /app
COPY --from=builder /app/server .
COPY index.html index.html
EXPOSE 8080
CMD ["./server"]
