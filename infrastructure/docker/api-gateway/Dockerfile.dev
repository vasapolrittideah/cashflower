FROM alpine
WORKDIR /app

ADD shared shared
ADD build build

ENTRYPOINT ["/app/build/api-gateway"]
