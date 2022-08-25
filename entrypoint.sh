#!/bin/bash

set -o pipefail  ## trace ERR through pipes
set -o errtrace  ## trace ERR through 'time command' and other functions
set -o nounset   ## set -u : exit the script if you try to use an uninitialised variable
set -o errexit   ## set -e : exit the script if any statement returns a non-true return value

if [[ "$#" -ge 1 ]] && [[ "$1" == install* ]]; then
    if [[ "$#" -eq 2 ]]; then
        case "$1" in
            "install"|"install.linux")
                echo "install helm-azure-tpl (linux) to $2"
                cp -- /helm-azure-tpl "$2/helm-azure-tpl"
                chmod +x "$2/helm-azure-tpl"
                exit 0
                ;;

            "install.windows")
                echo "install helm-azure-tpl (windows) to $2"
                cp -- /helm-azure-tpl.exe "$2/helm-azure-tpl.exe"
                exit 0
                ;;

            "install.darwin")
                echo "install helm-azure-tpl (darwin) to $2"
                cp -- /helm-azure-tpl.darwin "$2/helm-azure-tpl"
                chmod +x "$2/helm-azure-tpl"
                exit 0
                ;;

            *)
                >&2 echo "failed to install helm-azure-tpl: \"$1\" is not a valid install command"
                exit 1
                ;;
        esac
    else
        >&2 echo "failed to install helm-azure-tpl: target path not specified as argument"
        exit 1
    fi
fi

exec /helm-azure-tpl "$@"
