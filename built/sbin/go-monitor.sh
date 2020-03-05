#!/usr/bin/env bash


filepath=$(cd "$(dirname "$0")"; pwd);cd ${filepath}
source /etc/rc.d/init.d/functions
source /etc/profile

# 程序基本信息配置
PROCESS=process-name
PROGPATH="${filepath}/.."
LOGPATH="${PROGPATH}/log"
PROGRAM="${PROGPATH}/bin/${PROCESS}"
CONFIG="${PROGPATH}/conf/application.yml"

# 程序启动配置
STARTPROG="$PROGRAM -config $CONFIG"

# 配置文件集，用于check函数
FILES=(${CONFIG} )

# 微信告警对应责任人配置
wx_user="lvfei,xupeng,zhujin,fangwendong"


function send_wxmsg(){
    # subject="${PORCESS} DOWN, RESTART ALARM"
    msg="${PROCESS} DOWN, RESTART ALARM, from host: ${HOSTNAME}"
    # python ${filepath}/wechat_alter.py -m "${msg}" -u "${wx_user}"
    echo
}


function check(){
    echo -n "Checking ${PROCESS}: "
    cd $(dirname $0)
    while true; do
        for file in ${FILES[*]}; do
            if [ ! -f ${file} ]; then
                failure
                echo
                echo -n "${file}: no such file"
                break 2
            fi
        done

        mkdir -p ../log
        mkdir -p ../log_backup
        # mkdir -p ../stat
        success
        echo
        break
    done
}


function start(){
    check

    umask 022
    echo -n "Starting ${PROCESS}: "
    ulimit -c unlimited
    ulimit -HSn 655350

    cd ${filepath}
    DDD=`/bin/date +%Y-%m-%d--%H:%M:%S`
    mkdir -p ../log_backup/log_${DDD}
    mv ../log/std* ../log_backup/log_${DDD}/
    cd ../bin
    `nohup ${STARTPROG} 1>../log/stdout.log 2>../log/stderr.log &` && success || failure
    echo
}

function stop(){
    echo -n $"Stopping ${PROCESS}: "

    if [ -n "`pidofproc ${PROCESS}`" ] ; then
        killproc ${PROCESS}
    else
        progm="${PROCESS}: master"
        echo -n " -> \"$progm\": "
        if [ -n `pidofproc \"$progm\"` ] ; then
            killproc "$progm"
        else
            failure "Stopping ${PROCESS}"
        fi
    fi

    echo
}


function restart(){
        stop
        start
}


function monitor(){
    echo -n $"Monitoring ${PROCESS}: "
    if [ -n "`pidofproc ${PROCESS}`" ] ; then
            success $"Monitoring ${PROCESS}"
            echo
    else
        progm="${PROCESS}: master"
        echo -n " -> \"$progm\": "
        if [ -n "`pidofproc \"$progm\"`" ] ; then
            success $"Monitoring $progm"
            echo
        else
            warning $"Monitoring ${PROCESS} or $progm is not running, starting..."
            echo
            start
            send_wxmsg
        fi
    fi
}


case $1 in
    check)
        check
        ;;
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    monitor)
        monitor
        ;;
    *)
        echo "Usage:./go-monitor.sh check|start|stop|restart|monitor"
        exit 1
esac