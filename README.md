# graftd

## 0. Environment Info
System: Ubuntu 21.04
Docker Version: 20.10.12
Go Version: 1.17.8

## 1. Build images

We have 4 images to build, 1 image to pull:
```shell
# name: raft-demo 
# [a distributed kv-store]
# file:/graftd/Dockerfile 

# name: raft-client 
# [a kv-store client with curl]
# /graftd/client/Dockerfile

# name: logstash-http 
# [a pipeline that forward logs from log-pilot to custom vizor]
# /graftd/observer/logstash/Dockerfile

# name: vizor
# [a basic log visualization in form of table]
# /graftd/observer/receiver/Dockerfile

# name: log-pilot
# [a tool that can gather multiple containers' log]
# registry.cn-hangzhou.aliyuncs.com/acs/log-pilot:0.9.5-filebeat
```

you can build these by running:

```shell
cd ./scripts
./build.sh
```

## 2. Run demo

create a network named mynet first:

```shell
sudo docker network create --driver bridge --subnet 192.168.0.0/16 --gateway 192.168.0.1 mynet
```

### 2.1. start log vizor

**how we implement vizor?**

we need to modify our `hashicorp/raft` source code to add more custom logs:

```bash
go mod vendor
# to build with local vendor 
go build -mod vendor
```

logs aggr tools: https://github.com/AliyunContainerService/log-pilot

```bash
container_1 stdout-log ----
                    |
container_2 stdout-log ------ log-pilot --> (logstash + http-plugin) --> vizor as consumer
                    |
container_x stdout-log ----
```

**how to view logs?**

```shell
./build.sh
./run.sh
# use browser to open http://localhost:8090
# run node(s) or stop node(s)
# view logs
# if you want to stop
./stop.sh
```

### 2.2. start nodes

we recommend you to start 3 nodes to observer nodes' behaviours.

start graft node(s):
```shell
# start leader
sudo docker run -it --rm -P --net mynet \
    --cap-add=NET_ADMIN \
    --name node0 -h node0 \
    --label aliyun.logs.catalina=stdout \
    raft-demo /raftnode -id node0 ~/node0
# wait until node0 become leader
# start node1
sudo docker run -it --rm -P --net mynet \
    --cap-add=NET_ADMIN \
    --name node1 -h node1 \
    --label aliyun.logs.catalina=stdout \
    raft-demo /raftnode -id node1 -join node0:11000 ~/node1
# start node2
sudo docker run -it --rm -P --net mynet \
    --cap-add=NET_ADMIN \
    --name node2 -h node2 \
    --label aliyun.logs.catalina=stdout \
    raft-demo /raftnode -id node2 -join node0:11000 ~/node2
```

now back to browser to see logs.

### 2.3. interact with nodes

#### 2.3.1 kv-store client

you can start client like:
```shell
sudo docker run -it --rm --name raft-cli -h raft-cli --net mynet raft-client bash
```

make request (inside client):
```
root@raft-cli:/# curl -XPOST http://node0:11000/key -d '{"user":"kj"}'
root@raft-cli:/# curl http://node0:11000/key/user                     
{"user":"kj"}
```

and you can see log replication from browser vizor.

#### 2.3.2 chaos test

we can run chaos test in these nodes like:

```shell
# get into shell or run tc directly
sudo docker exec -it node0 /bin/bash
# delay
tc qdisc add dev eth0 root netem delay 10000ms
# or other network emulation powered by tc netem
```

## Q&A

### why `func GetLocalIP() string` is needed?

refs: https://github.com/hashicorp/raft/issues/438

> The node listening on 0.0.0.0:12322 needs to tell its peers how to reach it, i.e. its address. If the address is unspecified, we don't have a good way of knowing which of potentially many interfaces the user is expecting Raft to communicate on. So yes, this is a deliberate part of the design.

And we don't want to specify this address(raftBind Address) manually as we want to simplify the demo, so get into the containner and check local ip and then use it to start raft-node is not a good idea. In fact, we can do this inside go program (see `GetLocalIP()`).

### get warn like 'previous log not found' ?

refs: https://github.com/hashicorp/raft/issues/280

> What I saw happening was that when the new node joins it becomes a follower and then this [test](https://github.com/hashicorp/raft/blob/master/raft.go#L1072-L1075) is ran on the follower to check for conflicting log entries. This is a natural function of Raft which is used in log replication, where the leader checks the previous logs before appending any new entries. This is useful for when a follower may leave and rejoin the cluster with stale state. In this case, with a new follower, it is checking the follower’s log entries against the leader’s current log before appending any entries to the follower’s log. It sees there are no previous log entries since the follower is brand new. This triggers sending all the previous logs to the new follower. This is expected with new nodes, and not something to worry about!
