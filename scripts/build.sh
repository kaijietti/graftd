cd ../client
sudo docker build --tag graftd-client .
cd ../observer
cd ./logstash
sudo docker build --tag logstash-http .
cd ..
cd ./receiver
sudo docker build --tag vizor .
cd ../../
sudo docker build --tag graftd .
sudo docker pull registry.cn-hangzhou.aliyuncs.com/acs/log-pilot:0.9.5-filebeat