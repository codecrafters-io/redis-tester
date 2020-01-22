FROM alpine

RUN apk add go

# Python
RUN apk add python3=3.8.1-r0
RUN ln -s /usr/bin/python3 /usr/bin/python
RUN ln -s /usr/bin/pip3 /usr/bin/pip

ARG logstream_version=v14

ADD https://github.com/codecrafters-io/logstream/releases/download/${logstream_version}/${logstream_version}_linux_amd64.tar.gz /tmp/logstream.tar.gz
RUN tar -xvf /tmp/logstream.tar.gz -C /bin

ADD . /src
WORKDIR /src

RUN go build -o /usr/bin/redis-tester

CMD logstream -url=$LOGSTREAM_URL run redis-tester --binary-path=/app/spawn_redis_server.sh --config-path=/app/codecrafters.yml
