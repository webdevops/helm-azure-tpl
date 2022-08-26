#!/bin/sh
set -e

HELM_AZURE_TPL_VERSION=$(sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' "${HELM_PLUGIN_DIR}/plugin.yaml")

HOST_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
HOST_ARCH=$(uname -m | tr '[:upper:]' '[:lower:]')

case "$HOST_ARCH" in
    "x86_64")
        ## translate to amd64
        HOST_ARCH="amd64"
esac

PLUGIN_DOWNLOAD_FILE="helm-azure-tpl.${HOST_OS}.${HOST_ARCH}"
PLUGIN_DOWNLOAD_URL="https://github.com/webdevops/helm-azure-tpl/releases/download/${HELM_AZURE_TPL_VERSION}/${PLUGIN_DOWNLOAD_FILE}"
PLUGIN_TARGET_PATH="${HELM_PLUGIN_DIR}/${PLUGIN_DOWNLOAD_FILE}"

echo "download from: $PLUGIN_DOWNLOAD_URL"
echo "target path: $PLUGIN_TARGET_PATH"

curl --location "$PLUGIN_DOWNLOAD_URL" -o "$PLUGIN_TARGET_PATH"
if [ "$?" -nq 0 ];
    >&2 echo "[ERROR] failed to download plugin binary"
    exit 1
fi
chmod +x "$PLUGIN_TARGET_PATH"
