cd ..\client
docker build --tag graftd-client .
cd ..\observer
cd .\logstash
docker build --tag logstash-http .
cd ..
cd .\receiver
docker build --tag vizor .
cd ..\..\
docker build --tag graftd .
docker pull registry.cn-hangzhou.aliyuncs.com/acs/log-pilot:0.9.5-filebeat