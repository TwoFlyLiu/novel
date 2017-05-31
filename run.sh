#! /bin/bash
# Description: @description@
# Author: twoflyliu
# Mail: twoflyliu@163.com
# Create time: 2017 05 28 08:08:52

# 这个脚本是开发人员使用的, 最终的应用的是novel/app.py
# 需要安装python3, 和pygobject库
# 可以直接使用/distrib目录的debian包，内部使用golang和gtk+开发的
# golang开发后端， gtk+开发前缀，并且gnome桌面效果最好
#
# 理论上可以移植到windows上，但是显示效果不好，依赖包库超级多，
# 后面会推出不适用gtk+，直接使用win32 + duilib来进行开发

SCRIPT_DIR=${0%/*}
PWD_DIR=$(pwd)

run_go() {
    local obj=$1
    local executed_file="./$obj"
    shift
    cd $obj
    if [[ -x "$executed_file" ]]; then
        "$executed_file" $@
    else
        echo "build $obj ..."
        go build
        result=$?
        if [[ $result -eq 0 ]]; then
            $executed_file $@
        fi
    fi
}

run_py() {
    local obj=$1
    local executed_file="./app.py"
    shift
    cd $obj
    /usr/bin/python3 $executed_file $@
}

cd $SCRIPT_DIR
case $1 in
    'search')
        run_go $@
        ;;
    "novel")
        run_py $@
        ;;
    'backend')
        run_go $@
        ;;
esac
cd $PWD_DIR

exit 0

