FROM golang:1.20.3-alpine3.17 as builder

WORKDIR /app

COPY . /app

RUN go build -o myapp .

CMD ["./app/myapp"]