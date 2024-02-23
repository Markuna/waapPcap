需要在 linux 环境下编译运行
安装 go 1.21.2 环境

安装libpcap等环境依赖
yum install gcc
yum install libpcap-devel

设置cgo开关
go env -w CGO_ENABLED=1

挑选一个比较好的goProxy代理地址，加速下载依赖
1. 七牛 CDN
go env -w  GOPROXY=https://goproxy.cn,direct
2. 阿里云
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
3. 官方
go env -w  GOPROXY=https://goproxy.io,direct

在项目当前路径下执行

1、下载依赖

go mod tidy

2、编译

go build

3、在当前路径得到可执行二进制文件

./waapPcap

运行示例：

./waapPcap -i ens1f1 -t http://192.168.1.1:8080 -f host 192.168.123.111 and dst port 9106 

参数说明： -i 指定网卡 ，-f 设置过滤条件（和tcpdump一致），-t 设置waap防护地址

