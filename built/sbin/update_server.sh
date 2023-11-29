# 部署在工程目录中后：
# 输出： sh update_server.sh xxxserver.tar.gz.2023112409xx
# 下载服务 重启 测试(测试不通过需要另行手动回退服务)

set -e

if((1 == $#))
then
    server_dir=`echo $1 | awk -F '.' '{print $1}'`
    # ftpcli download yinhailong/$1 .
    wget xb-svr-ftp.oss-cn-hongkong.aliyuncs.com/$1
    tar zxvf $1
    ./${server_dir}/sbin/go-monitor.sh restart
    sleep 2
    ./${server_dir}/bin/httptest
    if [ $? -eq 0 ]; then
        echo "succeed"
    else
        echo "error"
    fi
fi


