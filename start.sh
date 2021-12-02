#!/bin/bash
set -e
if [ $# != 1 ] ; then
echo -e "\033[31m error:You can only enter one parameter, please use (bash start.sh -h) for help \033[0m"
exit 1;
fi
count=`ps aux | grep server\ serve | grep -v grep| wc -l`
function check(){
case $1 in
start)
   if [ 0 == $count ];then
     function_start
   else
     echo -e "\033[31m The service is running, Please shut down the service first or use (bash start.sh restart) \033[0m"
   fi
   ;;
stop)
  if [ 1 == $count ];then 
      function_stop
  else
     echo -e "\033[31m The service is not running, No need to stop \033[0m"
  fi
  ;;
restart)
    if [ 0 == $count ];then 
       function_start
    else
       function_restart
    fi
  ;;
update)
  pid=`ps aux | grep server\ serve | grep -v grep | awk -F ' ' '{print $2}'`
  echo -e "\033[32m now reload whiltlist or blacklist \033[0m"
  kill -SIGUSR1 $pid
  ;;
-h|--help)
  function_usage
  ;;
*)
  echo -e "\033[31m not support parameter:{$1} \033[0m"
  function_usage
esac
}
function_start(){
    echo -e "\033[32m start... \033[0m"
    nohup ./server serve -c ../conf/confile 2<&1 & 
    echo -e "\033[32m server is running... \033[0m"
    sleep 0.1
}
function_stop(){
   echo -e "\033[32m stop... \033[0m"
   ps aux | grep server\ serve | grep -v grep | awk -F ' ' '{print $2}' | xargs kill -SIGKILL
   echo -e "\033[32m server is stop... \033[0m"
}
function_restart(){
   function_stop
   sleep 1
   function_start
}
function_usage(){
   echo -e "\033[31m you should use parameter start|stop|restart|update. for example: bash start.sh start \033[0m"
   echo -e "\033[31m start:启动服务 \n stop:关闭服务\n restart:重启服务\n update:服务自动重新加载白名单(如果修改了白名单，只需执行此命令即可，无需重启程序) \033[0m"
}
check $1
