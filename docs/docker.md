# Docker

## Install docker
```shell
sudo apt-get update
curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun
```

## Build image
```shell
  sudo ./build_docker.sh
```

## Run
```shell
  sudo docker run -it -v /home/user/config:/karst -e INIT_ARGS="-c /karst/config.json" --name test-container --network host crustio/karst:0.2.0
```