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
wget -O 'built/go/env_check.py' https://github.com/xbonlinenet/goup/raw/master/built/go/env_check.py
wget -O 'built/make/common.make' https://github.com/xbonlinenet/goup/raw/master/built/make/common.make
wget -O 'built/make/init.make' https://github.com/xbonlinenet/goup/raw/master/built/make/init.make
wget -O 'built/sbin/go-monitor.sh' https://github.com/xbonlinenet/goup/raw/master/built/sbin/go-monitor.sh
wget -O 'built/version' https://github.com/xbonlinenet/goup/raw/master/built/version

wget -O 'conf/dev/data.yml' https://github.com/xbonlinenet/goup/raw/master/conf/product/data.yml
wget -O "conf/dev/$module.yml" https://github.com/xbonlinenet/goup/raw/master/conf/product/demo.yml
$sed_cmd "s/0.0.0.0:13360/0.0.0.0:$port/g"  conf/dev/$module.yml
$sed_cmd -i "s/demo-application/$parent\/$module/g"  conf/dev/$module.yml


wget -O "pkg/cmd/$module/main.go" https://github.com/xbonlinenet/goup/raw/master/main.go
$sed_cmd -i "s/github.com\/xbonlinenet\/goup\/demo/coding.xbonline.net\/$parent\/$module\/pkg\/cmd\/$module\/api\/demo/g" pkg/cmd/$module/main.go

wget -O "pkg/cmd/$module/api/demo/echo.go" https://github.com/xbonlinenet/goup/raw/master/demo/echo.go
wget -O "pkg/cmd/$module/api/demo/redis.go" https://github.com/xbonlinenet/goup/raw/master/demo/redis.go
wget -O "pkg/cmd/$module/api/demo/mysql.go" https://github.com/xbonlinenet/goup/raw/master/demo/mysql.go
wget -O "pkg/cmd/$module/api/demo/config.go" https://github.com/xbonlinenet/goup/raw/master/demo/config.go
wget -O "pkg/cmd/$module/api/demo/pre.go" https://github.com/xbonlinenet/goup/raw/master/demo/pre.go


wget -O "pkg/cmd/$module/httptest/echo_test.go" https://github.com/xbonlinenet/goup/raw/master/httptest/echo_test.go
$sed_cmd "s/127.0.0.1:13360/127.0.0.1:$port/g"  conf/dev/$module.yml


cat << EOF > go.mod
module coding.xbonline.net/$server

go 1.12

replace (
        cloud.google.com/go => github.com/googleapis/google-cloud-go v0.26.0

        go.uber.org/atomic => github.com/uber-go/atomic v1.4.0
        go.uber.org/multierr => github.com/uber-go/multierr v1.1.0
        go.uber.org/zap => github.com/uber-go/zap v1.10.0

        golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190605123033-f99c8df09eb5
        golang.org/x/net => github.com/golang/net v0.0.0-20190607181551-461777fb6f67
        golang.org/x/sync => github.com/golang/sync v0.0.0-20190423024810-112230192c58
        golang.org/x/sys => github.com/golang/sys v0.0.0-20190610200419-93c9922d18ae
        golang.org/x/text => github.com/golang/text v0.3.2
        google.golang.org/appengine => github.com/golang/appengine v1.6.1
        gopkg.in/fsnotify.v1 => github.com/fsnotify/fsnotify v1.4.2
        gopkg.in/jcmturner/gokrb5.v6 => github.com/jcmturner/gokrb5 v6.0.5+incompatible
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