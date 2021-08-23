FROM alpine:3.10

ADD bin/gateway /opt/gmqtt/gateway
ADD conf/conf.toml /opt/gmqtt/conf/conf.toml

EXPOSE 1883
EXPOSE 8080
WORKDIR /opt/gmqtt

ENTRYPOINT [ "./gateway" ]
