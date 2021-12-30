# logs viz

## build images

logstash (with http output):

```
cd ./logstash
sudo docker build --tag logstash-http .
```

receiver (TODO):

```
cd ./receiver
sudo docker build --tag vizor .
```

## run 

start receiver:
```
sudo docker run -it --rm --name vizor --net mynet -P vizor /vizor
```

start logstash:
```
sudo docker run -it -P --name logstash-http -h logstash-http --net mynet logstash-http
```

start log-pilot:
```
sudo docker run --rm -it -P --net mynet \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /etc/localtime:/etc/localtime \
    -v /:/host:ro \
    --cap-add SYS_ADMIN \
    -e LOGGING_OUTPUT=logstash \
    -e LOGSTASH_HOST=logstash-http \
    -e LOGSTASH_PORT=5044 \
    registry.cn-hangzhou.aliyuncs.com/acs/log-pilot:0.9.5-filebeat
```

start graft node(s):
```
sudo docker run -it --rm -P --name node0 -h node0 --net mynet --label aliyun.logs.catalina=stdout  raft-demo /raftexample -id node0 ~/node0

sudo docker run -it --rm -P --name node1 -h node1 --net mynet --label aliyun.logs.catalina=stdout  raft-demo /raftexample -id node1 ~/node1
```