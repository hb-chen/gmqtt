# Micro MQ
以微服务+MQ构建支持高并发连接的分布式消息服务系统

![micro-mq](/doc/img/architecture.jpg "micro-mq")

### Components
- gateway
- auth
- broker
- console

### Frameworks
- [rpcx](https://github.com/smallnest/rpcx)
- [Echo](https://github.com/labstack/echo)

### 服务依赖
- 服务注册与发现
    - Etcd
    - Zookeeper
    - Consul
- MQ
    - Kafka

## CMD
### Protobuf
```bash
protoc -I=. \
  --go_out=. \
  auth/proto/auth.proto
```

## 参考内容
- [[译]百万级WebSockets和Go语言](http://xiecode.cn/post/cn_06_a_million_websockets_and_go/)
    - [[原]A Million WebSockets and Go](https://medium.freecodecamp.org/million-websockets-and-go-cc58418460bb)
- [SurgeMQ](https://github.com/surgemq/surgemq)
- [Micro](http://github.com/micro)
- [Managing IoT devices with Kafka and MQTT](https://www.ibm.com/blogs/bluemix/2017/01/managing-iot-devices-with-kafka-and-mqtt/)