#!/bin/bash
# 参数: 模块名称， 端口

echo '欢迎使用小步网络 golang 服务脚手架工具'
echo '请按引导进行操作'




echo '第一步：输入项目的名称。和 git 代码仓库名字一致， 比如 https://coding.xbonline.net/server/config_center， 这输入 server/config_center'

read server


items=`echo $server | tr '/' ' '`
array=($items)
if [ ${#array[@]} -ne 2 ];
then
        echo "项目名称格式非法， 合法示例: server/test. 必须有 / 分割"
        exit 2
else
        echo ''
fi

parent=${array[0]}
module=${array[1]}


if [ -d "$module" ]; then
        echo "$module 目录已经存在不能初始化"
        exit 1
fi


echo '第二步：输入服务端口号。端口范围 > 15000'

read port


echo '正常创建目录结构...'
mkdir -p $module

cd $module

mkdir -p built/go
mkdir -p built/make
mkdir -p built/sbin

mkdir -p conf/dev
mkdir -p conf/test
mkdir -p conf/product
mkdir -p pkg/cmd/$module/api/demo
mkdir -p pkg/cmd/$module/httptest


if [ "$(uname)" == "Darwin" ]; then
    sed_cmd="sed -i \'\'"
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
    sed_cmd="sed -i "
elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW32_NT" ]; then
    sed_cmd="sed -i "
elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW64_NT" ]; then
    sed_cmd="sed -i "
fi


echo '现在开始下载必须的文件...'

wget -O '.gitignore' https://github.com/xbonlinenet/goup/raw/master/.gitignore
wget -O 'Makefile' https://github.com/xbonlinenet/goup/raw/master/Makefile
$sed_cmd -i "s/module=demo/module=$module/g"  Makefile
$sed_cmd -i "s/group=group/group=$parent/g"  Makefile

wget -O 'built/go/env_check.py' https://github.com/xbonlinenet/goup/raw/master/built/go/env_check.py
wget -O 'built/make/common.make' https://github.com/xbonlinenet/goup/raw/master/built/make/common.make
wget -O 'built/make/init.make' https://github.com/xbonlinenet/goup/raw/master/built/make/init.make
wget -O 'built/sbin/go-monitor.sh' https://github.com/xbonlinenet/goup/raw/master/built/sbin/go-monitor.sh
wget -O 'built/sbin/update_server.sh' https://github.com/xbonlinenet/goup/raw/master/built/sbin/update_server.sh
wget -O 'built/version' https://github.com/xbonlinenet/goup/raw/master/built/version

wget -O 'conf/dev/data.yml' https://github.com/xbonlinenet/goup/raw/master/conf/product/data.yml
wget -O "conf/dev/$module.yml" https://github.com/xbonlinenet/goup/raw/master/conf/product/demo.yml
$sed_cmd "s/0.0.0.0:13360/0.0.0.0:$port/g"  conf/dev/$module.yml
$sed_cmd -i "s/demo-application/$parent\/$module/g"  conf/dev/$module.yml


wget -O "pkg/cmd/$module/main.go" https://github.com/xbonlinenet/goup/raw/master/pkg/cmd/demo/main.go
$sed_cmd -i "s/github.com\/xbonlinenet\/goup\/demo/coding.xbonline.net\/$parent\/$module\/pkg\/cmd\/$module\/api\/demo/g" pkg/cmd/$module/main.go

wget -O "pkg/cmd/$module/api/demo/echo.go" https://github.com/xbonlinenet/goup/raw/master/demo/echo.go
wget -O "pkg/cmd/$module/api/demo/redis.go" https://github.com/xbonlinenet/goup/raw/master/demo/redis.go
wget -O "pkg/cmd/$module/api/demo/mysql.go" https://github.com/xbonlinenet/goup/raw/master/demo/mysql.go
wget -O "pkg/cmd/$module/api/demo/config.go" https://github.com/xbonlinenet/goup/raw/master/demo/config.go
wget -O "pkg/cmd/$module/api/demo/pre.go" https://github.com/xbonlinenet/goup/raw/master/demo/pre.go
wget -O "pkg/cmd/$module/api/demo/sleep.go" https://github.com/xbonlinenet/goup/raw/master/demo/sleep.go
wget -O "pkg/cmd/$module/api/demo/doc.go" https://github.com/xbonlinenet/goup/raw/master/demo/doc.go
wget -O "pkg/cmd/$module/api/demo/struct.go" https://github.com/xbonlinenet/goup/raw/master/demo/struct.go
wget -O "pkg/cmd/$module/api/demo/struct.go" https://github.com/xbonlinenet/goup/raw/master/demo/postgres.go
wget -O "pkg/cmd/$module/api/demo/struct.go" https://github.com/xbonlinenet/goup/raw/master/demo/react.go


wget -O "pkg/cmd/$module/httptest/echo_test.go" https://github.com/xbonlinenet/goup/raw/master/httptest/echo_test.go
$sed_cmd "s/127.0.0.1:13360/127.0.0.1:$port/g"  conf/dev/$module.yml


cat << EOF > go.mod
module coding.xbonline.net/$server

go 1.13

require (
	github.com/bxcodec/faker/v3 v3.5.0
	github.com/gavv/httpexpect v2.0.0+incompatible
	github.com/gin-gonic/gin v1.7.7
	github.com/jinzhu/gorm v1.9.16
	github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/onsi/gomega v1.9.0 // indirect
	github.com/xbonlinenet/goup v0.3.0
	go.uber.org/zap v1.19.1
)

EOF

mkdir test
cat << EOF > test/$module.py

import os
import sys

dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

ret = os.system(dir + "/bin/httptest")
if ret != 0:
    print("HTTP API Test Not Success")
    sys.exit(1)
EOF


echo "现在通过 Makefile 编译运行你的服务吧"
echo "make clean 清空 编译的缓存及临时文件"
echo "make build module=$module env=dev 编译你的服务"
echo "make build run module=$module env=dev 编译并运行你的服务"
echo "make clean release module=$module env=product 将会打包一个生产环境的包，一般情况下你不会使用到"

echo "现在开始吧~~~~"
echo "#######         cd $module;  make build run module=$module env=dev                     ######"
echo "####### 在浏览器打开地址 http://localhost:$port/doc/list 查看接口列表     #######"
echo "####### 在浏览器打开地址 http://localhost:$port/metrics 查看prometheus监控项     #######"