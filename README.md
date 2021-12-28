# graftd

## build demo 

build raft-backend key-value store:

```shell
git clone https://github.com/kaijietti/graftd.git
cd ./graftd/raftexample
sudo docker build --tag raft-demo .
```

build client(TODO: cli now is simply a curl tool):

```shell
cd ./graftd/client
sudo docker build --tag raft-client .
```

## run demo

create a network first:

```shell
sudo docker network create --driver bridge --subnet 192.168.0.0/16 --gateway 192.168.0.1 mynet
```

### bootstrap fisrt node

start a leader node:

```shell
sudo sudo docker run -it -P --name nodew -h nodew --net mynet raft-demo /raftexample -id nodew ~/nodew
```

start client:
```shell
sudo docker run -it --name raft-cli -h raft-cli --net mynet raft-client bash
```

make request (inside client):
```shell
root@raft-cli2:/# curl -XPOST http://noder:11000/key -d '{"user":"kj"}'
root@raft-cli2:/# curl http://noder:11000/key/user                     
{"user":"kj"}
```

### make cluster

key: use `join` parameter to join a new follower node to a leader node

```shell
sudo sudo docker run -it -P --name nodee -h nodee --net mynet raft-demo /raftexample -id nodee -join nodew:11000 ~/nodee

sudo sudo docker run -it -P --name noder -h noder --net mynet raft-demo /raftexample -id noder -join nodew:11000 ~/noder
```

## test something

### test node crash

TODO

### test split brain

TODO


## Q&A

### why `func GetLocalIP() string` is need?

refs: https://github.com/hashicorp/raft/issues/438

> The node listening on 0.0.0.0:12322 needs to tell its peers how to reach it, i.e. its address. If the address is unspecified, we don't have a good way of knowing which of potentially many interfaces the user is expecting Raft to communicate on. So yes, this is a deliberate part of the design.

And we don't want to specify this address(raftBind Address) manually as we want to simplify the demo, so get into the containner and check local ip and then use it to start raft-node is not a good idea. In fact, we can do this inside go program (see `GetLocalIP()`).