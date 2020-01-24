FROM alpine

RUN apk add go
RUN apk add curl

# Python
RUN apk add python3=3.8.1-r0
RUN ln -s /usr/bin/python3 /usr/bin/python
RUN ln -s /usr/bin/pip3 /usr/bin/pip
RUN pip install pipenv

ADD . /src
WORKDIR /src

RUN go build -o /usr/bin/redis-tester

WORKDIR /app

ENV APP_DIR=/app
CMD redis-tester
