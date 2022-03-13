sudo docker run --rm -itd --name vizor -p 8090:8090 vizor /vizor
sudo docker network connect mynet vizor
sudo docker run --rm -itd -P --name logstash-http -h logstash-http --net mynet logstash-http
sudo docker run --rm -itd -P --name log-pilot --net mynet -v /var/run/docker.sock:/var/run/docker.sock -v /etc/localtime:/etc/localtime -v /:/host:ro --cap-add SYS_ADMIN -e LOGGING_OUTPUT=logstash -e LOGSTASH_HOST=logstash-http -e LOGSTASH_PORT=5044 registry.cn-hangzhou.aliyuncs.com/acs/log-pilot:0.9.5-filebeat
