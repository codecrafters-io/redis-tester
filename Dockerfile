FROM alpine

RUN apk add go

ARG logstream_version=v14

ADD https://github.com/codecrafters-io/logstream/releases/download/${logstream_version}/${logstream_version}_linux_amd64.tar.gz /tmp/logstream.tar.gz
RUN tar -xvf /tmp/logstream.tar.gz -C /bin

ADD . /src
WORKDIR /src

RUN go build -o /usr/bin/redis-tester

CMD logstream -url=$LOGSTREAM_URL run redis-tester --binary-path=/app/spawn_redis_server.sh --config-path=/app/.codecrafters.yml
