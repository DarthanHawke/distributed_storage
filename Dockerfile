FROM golang:1.23 AS build

COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux go build -o kvs /src/cmd/distributed_storage

FROM scratch

COPY --from=build /src/kvs .
COPY --from=build /src/*.env .
COPY --from=build /src/config/*.yaml ./config/

EXPOSE 8080

CMD ["/kvs"]