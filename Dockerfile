FROM golang:1.17-alpine

ENV CGO_ENABLED=0

WORKDIR /go/src/github.com/21hack02win/nascalay-backend
COPY . .

RUN apk upgrade --update && \
    apk --no-cache add git

RUN go install github.com/cosmtrek/air@latest

CMD ["air", "-c", ".air.toml"]
