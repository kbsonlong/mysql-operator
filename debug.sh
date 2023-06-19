#! /bin/sh
###
 # @Author: kbsonlong kbsonlong@gmail.com
 # @Date: 2023-06-05 17:46:47
 # @LastEditors: kbsonlong kbsonlong@gmail.com
 # @LastEditTime: 2023-06-05 17:46:53
 # @FilePath: /mysql-operator/debug.sh
 # @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
### 

export GOPROXY=https://goproxy.cn
dlv --headless --log --listen :9009 --api-version 2 --accept-multiclient debug main.go