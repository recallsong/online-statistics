# online-statistics
统计tcp连接、websocket连接的连接时长和连接数的一个服务程序。

# Topology

![Topology](https://github.com/recallsong/online-statistics/raw/master/docs/img/topology.png)

# Demo

[Admin Page](http://test.songos.top:8610/admin/)

![AdminPage](https://github.com/recallsong/online-statistics/raw/master/docs/img/admin-page.png)

# Download

        go get github.com/recallsong/online-statistics

# Config
可以在配置文件中配置监听端口号、redis地址、keepalive等

    conf/config.yml

# Run

    make run

# Build To Docker Image

    make docker-build

# Run In Docker Container

    make docker-run

# License
[Apache License 2.0](https://github.com/recallsong/online-statistics/blob/master/LICENSE)