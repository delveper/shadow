FROM golang:1.20.3-alpine3.17 as build

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN scripts/setup.sh
RUN  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o myapp ./cmd/main.go

FROM alpine:3.17 as prod

COPY --from=build /app .

ENTRYPOINT ["./app/myapp"]