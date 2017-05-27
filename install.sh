#! /bin/bash
# Description: @description@
# Author: twoflyliu
# Mail: twoflyliu@163.com
# Create time: 2017 05 25 18:30:20

install() {
    case $1 in
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
        "all")
            install extracter
            install engine
            ;;
        *)
            echo unsupport install command!:$1
            echo "Usage: $0 [engine|extracter]"
            ;;
    esac
}

for obj in $@; do
    install ${obj}
done

exit 0
