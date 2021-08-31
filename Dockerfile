# build stage
FROM golang:alpine as builder

ENV GO111MODULE=on
WORKDIR /app
COPY go.* ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o main ./cmd/goerd/.

# create application stage
FROM alpine as app

RUN apk --no-cache upgrade && apk --no-cache add ca-certificates
COPY --from=builder /app/main /usr/local/bin/goerd
WORKDIR /goerd

CMD ["goerd"]
