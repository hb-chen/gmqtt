# Micro MQ
以微服务+MQ构建支持高并发连接的分布式消息服务系统

> [消息持久化及写放大模式设计](https://github.com/hb-go/micro-mq/issues/1#issuecomment-518517597)

![micro-mq](/doc/img/architecture.jpg "micro-mq")

- Gateway节点`Node`通过订阅MQ消息的方式，完成消息在节点间的转发
- 根据业务场景的需求，需要考虑Node节点消息消费与生产速度的匹配
    - Client间的pub/sub关系比较多，如`n`个node, 发送`m`条消息/topic/node，节点消费的需求是`n`*`m`条/topic
    - Client端多为pub操作，而系统下发消息较少的情况，节点消费需求则比较低
    
Gateway编程模型

![micro-mq](/doc/img/gateway_model.jpg "gateway_modem")
### Features

## 运行
#### 服务依赖
> 根据配置选择 conf/conf.toml
- MQ `conf.broker`
    - [x] Kafka 
- 服务注册与发现 `conf.auth`
    - [x] Etcd
    - [ ] Consul
    
#### 启动服务
```bash
# 启动Gateway，[-h]帮助查看可选参数
$ cd gateway
$ go run -tags "etcd" main.go

# RPC Auth服务，[-h]帮助查看可选参数
$ cd api
$ go run -tags "etcd" main.go
```

#### MQTT Web Client
[在线演示](http://mqtt-client.hbchen.com/)
> [Github源码](https://github.com/hb-chen/hivemq-mqtt-web-client)

## 组件
- gateway
    - sessions
        - [x] mock
        - [x] redis
    - topic
        - [x] mem
    - client auth
        - [x] mock
        - [x] rpc
    - pub/sub auth
        - [ ] mock
        - [ ] rpc
- broker
    - [x] kafka
    
- > Developing
    - api
        - auth
            - [ ] RPC服务间的访问控制：RBAC
        - client
            - auth
                - [x] client auth
                - [ ] pub/sub auth
            - register
                - [ ] register
                - [ ] unregister
        - cluster
            - nodes
                - 节点列表
            - clients
                - 终端列表
            - sessions
                - 会话列表
            - topics
                - 话题信息
            - subscriptions
                - 订阅信息
    - console
    - deploy

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
$ go test -run=TestClients$
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
# 在.proto有import时注意相对的路径
protoc -I=$GOPATH/src:.  --go_out=.  api/proto/define.proto
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
- [[转][译]百万级WebSockets和Go语言](https://colobu.com/2017/12/13/A-Million-WebSockets-and-Go/)
    - [[原]A Million WebSockets and Go](https://medium.freecodecamp.org/million-websockets-and-go-cc58418460bb)
- [SurgeMQ](https://github.com/surgemq/surgemq)
- [Micro](http://github.com/micro)
- [Managing IoT devices with Kafka and MQTT](https://www.ibm.com/blogs/bluemix/2017/01/managing-iot-devices-with-kafka-and-mqtt/)
