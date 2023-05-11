#!/bin/sh
###
 # @Author: kbsonlong kbsonlong@gmail.com
 # @Date: 2023-05-10 22:26:56
 # @LastEditors: kbsonlong kbsonlong@gmail.com
 # @LastEditTime: 2023-05-11 13:14:11
 # @FilePath: /mysql-operator/scripts/init.sh
 # @Description: 初始化Pod
### 

Init_Server_ID() {
    cat << EOF > /etc/mysql/mysql.conf.d/dynamic.cnf
[mysqld]
## 注意要唯一
server-id=1$(echo $HOSTNAME | awk -F '-' '{print $NF}')
## 开启二进制日志功能
log-bin=mysql-bin
EOF
}

Init_Server_ID