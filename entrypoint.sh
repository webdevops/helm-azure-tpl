#!/bin/bash

set -o pipefail  ## trace ERR through pipes
set -o errtrace  ## trace ERR through 'time command' and other functions
set -o nounset   ## set -u : exit the script if you try to use an uninitialised variable
set -o errexit   ## set -e : exit the script if any statement returns a non-true return value

if [[ "$#" -ge 1 ]] && [[ "$1" == install* ]]; then
    if [[ "$#" -eq 2 ]]; then
        COMMAND="$1"
        TARGET_DIR="$2"
        case "$COMMAND" in
            "install"|"install.linux")
                echo "install helm-azure-tpl (linux) to $TARGET_DIR"
                if [ -d "$TARGET_DIR" ]; then
                    cp -- /helm-azure-tpl "$TARGET_DIR/helm-azure-tpl"
                    chmod +x "$TARGET_DIR/helm-azure-tpl"
                    exit 0
                else
                     >&2 echo "target directory \"$TARGET_DIR\" doesn't exists"
                    exit 1
                fi
                ;;

            *)
                >&2 echo "failed to install helm-azure-tpl: \"$COMMAND\" is not a valid install command"
                exit 1
                ;;
        esac
    else
        >&2 echo "failed to install helm-azure-tpl: target path not specified as argument"
        exit 1
    fi
fi

exec /helm-azure-tpl "$@"
