FROM golang:latest as builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -installsuffix cgo -o backup .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /app/backup .
EXPOSE 8080

CMD ["./backup"]
