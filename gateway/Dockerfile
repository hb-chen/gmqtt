FROM alpine:3.10

ADD bin/linux/gateway /opt/gmqtt/gateway
ADD config/gateway /opt/gmqtt/conf.d

EXPOSE 1883
EXPOSE 8080
WORKDIR /opt/gmqtt

ENTRYPOINT [ "./gateway" ]
