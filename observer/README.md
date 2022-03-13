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
sudo docker run --rm -it --name vizor --net mynet -P vizor /vizor
```

receiver bootstrap:

we need browser to view the log flow, but our host machine is not able to access container inside `mynet`.

so we can first start like this:

```
docker run --rm -it --name vizor -p 8090:8090 vizor /vizor
```

and then according to [add-containers-to-a-network](https://docs.docker.com/engine/tutorials/networkingcontainers/#add-containers-to-a-network):

```
docker network connect mynet vizor
```

start logstash:
```
sudo docker run --rm -it -P --name logstash-http -h logstash-http --net mynet logstash-http
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

docker run --rm -it -P --net mynet -v /var/run/docker.sock:/var/run/docker.sock -v /etc/localtime:/etc/localtime -v /:/host:ro --cap-add SYS_ADMIN -e LOGGING_OUTPUT=logstash -e LOGSTASH_HOST=logstash-http -e LOGSTASH_PORT=5044 registry.cn-hangzhou.aliyuncs.com/acs/log-pilot:0.9.5-filebeat
```

start graft node(s):
```
sudo docker run -it --rm -P --cap-add=NET_ADMIN --name node0 -h node0 --net mynet --label aliyun.logs.catalina=stdout  raft-demo /raftnode -id node0 ~/node0

sudo docker run -it --rm -P --cap-add=NET_ADMIN --name node1 -h node1 --net mynet --label aliyun.logs.catalina=stdout  raft-demo /raftnode -id node1 -join node0:11000 ~/node1 
```

now back to receiver to see logs.

## TODO

more awesome viz project: http://kanaka.github.io/raft.js/