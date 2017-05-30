#! /bin/bash
# Description: @description@
# Author: twoflyliu
# Mail: twoflyliu@163.com
# Create time: 2017 05 28 08:08:52

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

