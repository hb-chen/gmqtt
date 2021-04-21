FROM alpine:3.10

ADD bin/gateway /usr/share/gmqtt/gateway
ADD conf/conf.toml /usr/share/gmqtt/conf/conf.toml

WORKDIR /usr/share/gmqtt
ENTRYPOINT [ "gateway" ]
