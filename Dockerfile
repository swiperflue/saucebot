FROM golang AS build
WORKDIR /usr/src/go
ENV CGO_ENABLED 0
COPY * /usr/src/go/

RUN go build -o saucebot .
FROM alpine
COPY --from=build /usr/src/go/saucebot /usr/src/go/config.json /bot/
WORKDIR /bot
CMD ["./saucebot"]
