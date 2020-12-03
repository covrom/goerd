# build stage
FROM golang:1.15.5-alpine3.12 as builder

ENV GO111MODULE=on
WORKDIR /app
COPY go.* ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# create application stage
FROM alpine:3.10 as app

RUN apk --no-cache upgrade && apk --no-cache add ca-certificates
COPY --from=builder /app/goerd /usr/local/bin/goerd
WORKDIR /goerd

CMD ["goerd"]
