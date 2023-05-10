#!/bin/bash
###
 # @Author: kbsonlong kbsonlong@gmail.com
 # @Date: 2023-05-10 22:26:56
 # @LastEditors: kbsonlong kbsonlong@gmail.com
 # @LastEditTime: 2023-05-10 22:29:21
 # @FilePath: /mysql-operator/scripts/init.sh
 # @Description: 初始化Pod
### 

Init_Server_ID() {
    echo "[mysqld]" >/etc/mysql/mysql.conf.d/dynamic.cnf
    echo "server-id=$(echo $HOSTNAME | awk -F '-' '{print $NF}')" >> /etc/mysql/mysql.conf.d/dynamic.cnf
}

Init_Server_ID