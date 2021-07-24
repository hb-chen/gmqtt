## API

### 安全
- 服务间访问控制`auth`
    - Token认证
    - 
- 用户访问控制`access`
- 终端访问控制`client/auth`
    - Auth认证
    - 话题订阅/消费鉴权
    
### Gateway集群
- Nodes
    - summary
    - list
    - detail
    - > Etcd / Consul / ZooKeeper，同RPC
- Clients
    - summary
    - list
    - > MySQL / PostgreSQL / NoSQL
- Sessions
    - list
    - detail
    - topics
    - > Redis: (SortedSet)ClientId => CreateTime, (String)ClientId => session
- Topics
    - > Redis: (SortedSet)NodeId-Topic => QOS
- Subscriptions
    - > Redis: (SortedSet)ClientId-Topic => QOS
- Broker
    - consumer group `NodeId`
    - topics
        - offset
    - > kafka-manager