#! /bin/bash
# Description: @description@
# Author: twoflyliu
# Mail: twoflyliu@163.com
# Create time: 2017 05 25 18:30:20

SCRIPT_DIR=${0%/*}
PWD_DIR=$(pwd)

install() {
    case $1 in
        "tool")
            echo install tool...
            cd tool
            go install
            cd ..
            ;;
        "engine")
            echo install engine...
            cd engine
            go install
            cd ..
            ;;
        "extracter")
            echo install extracter...
            cd extracter
            go install
            cd ..
            ;;
        "search")
            echo build search...
            cd search
            go build 
            result=$?
            if [[ $result -eq '0' ]]; then
                mv ./search ../novel
            fi
            cd ..
            ;;
        "backend")
            echo build backend...
            cd backend
            go build
            result=$?
            if [[ $result -eq '0' ]]; then
                mv ./backend ../novel
            fi
            cd ..
            ;;
        "all")
            install tool
            install engine
            install extracter
            install search
            install backend
            ;;
        *)
            echo unsupport install command!:$1
            echo "Usage: $0 [engine|extracter]"
            ;;
    esac
}

cd $SCRIPT_DIR
for obj in $@; do
    install ${obj}
done
cd $PWD_DIR

exit 0
