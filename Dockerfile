FROM golang:1.17.4-alpine

WORKDIR /app
COPY ./ /app

RUN go mod init main \
  && go mod tidy \
  && go build

EXPOSE 3000


CMD ["go", "run", "main.go"]