# Micro MQ
以微服务+MQ构建支持高并发连接的分布式消息服务系统

![micro-mq](/doc/img/architecture.jpg "micro-mq")

## 运行
#### 服务依赖
- 服务注册与发现
    - [x] Etcd
    - [ ] Consul
- MQ
    - [x] Kafka
    
#### 启动服务
```bash
# 启动Auth服务，[-h]帮助查看可选参数
$ cd auth
$ go run -tags "etcd" main.go

# 启动Gateway，[-h]帮助查看可选参数
$ cd gateway
$ go run -tags "etcd" main.go
```

#### MQTT Web Client
[在线演示](http://mqtt-client.hbchen.com/)
> [Github源码](https://github.com/hb-chen/hivemq-mqtt-web-client)

## 组件
- [x] gateway
    - [x] sessions
        - [x] mem
        - [ ] redis
    - [x] topic
        - [x] mem
- [x] auth
    - [x] mock
    - [x] rpc
- [x] broker
    - [x] kafka
- [ ] console

## Frameworks
- [rpcx](https://github.com/smallnest/rpcx)
- [Echo](https://github.com/labstack/echo)

## Benchmark
- MBP开发环境
- mock broker
- pub timeout 3000 ms, size 64
```bash
$ go test -run=TestClient$
# 1 client pub, sent 100000 messages, no sub
# qos 0
Total sent 100000 messages dropped 0 in 3096.074512 ms, 0.030961 ms/msg, 32298 msgs/sec
# qos 1
Total sent 100000 messages dropped 0 in 10411.318733 ms, 0.104113 ms/msg, 9604 msgs/sec
```
```bash
$ go test -run=TestClient$
# 1000 clients, sent 1000 messages/client, no sub
# qos 0
Total Sent 1000000 messages dropped 0 in 14628.666153 ms, 0.014629 ms/msg, 68358 msgs/sec
# qos 1
Total Sent 1000000 messages dropped 0 in 38669.812430 ms, 0.038670 ms/msg, 25859 msgs/sec

# 1000 clients, sent 1000 messages/client, 1 client sub
# qos 1
Total Sent 1000000 messages dropped 0 in 65403.199238 ms, 0.065403 ms/msg, 15289 msgs/sec
Total Received 1000000 messages in 65403.199238 ms, 0.065403 ms/msg, 15289 msgs/sec
# qos 2
Total Sent 1000000 messages dropped 0 in 68339.624216 ms, 0.068340 ms/msg, 14632 msgs/sec
Total Received 1000000 messages in 68339.624216 ms, 0.068340 ms/msg, 14632 msgs/sec
```

## 备用命令
#### Protobuf
```bash
protoc -I=. \
  --go_out=. \
  auth/proto/auth.proto
```
#### Kafka
```bash
# Start the server
$ bin/zookeeper-server-start.sh config/zookeeper.properties
$ bin/kafka-server-start.sh config/server.properties

# List topic
bin/kafka-topics.sh --list --zookeeper localhost:2181

# Start a consumer
bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic topic --from-beginning
```
#### Kafka Manager
```bash
# 启动报错，需修改conf/application.conf
kafka-manager.zkhosts="localhost:2181"
```

## 参考内容
- [[译]百万级WebSockets和Go语言](http://xiecode.cn/post/cn_06_a_million_websockets_and_go/)
    - [[原]A Million WebSockets and Go](https://medium.freecodecamp.org/million-websockets-and-go-cc58418460bb)
- [SurgeMQ](https://github.com/surgemq/surgemq)
- [Micro](http://github.com/micro)
- [Managing IoT devices with Kafka and MQTT](https://www.ibm.com/blogs/bluemix/2017/01/managing-iot-devices-with-kafka-and-mqtt/)