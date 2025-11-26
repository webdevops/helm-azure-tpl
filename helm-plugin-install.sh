#!/bin/bash
set -e
set -o pipefail

HELM_AZURE_TPL_VERSION=$(sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' "${HELM_PLUGIN_DIR}/plugin.yaml")

HOST_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
HOST_ARCH=$(uname -m | tr '[:upper:]' '[:lower:]')

FILE_SUFFIX=""

case "${HOST_OS}" in
	cygwin*)
	    HOST_OS="windows"
	    FILE_SUFFIX=".exe"
	    ;;
	mingw*)
	    HOST_OS="windows"
	    FILE_SUFFIX=".exe"
	    ;;
esac

case "$HOST_ARCH" in
    "x86_64")
        ## translate to amd64
        HOST_ARCH="amd64"
        ;;
    "aarch64")
        ## translate to arm64
        HOST_ARCH="arm64"
        ;;
esac

PLUGIN_DOWNLOAD_FILE="helm-azure-tpl${FILE_SUFFIX}"
PLUGIN_DOWNLOAD_URL="https://github.com/webdevops/helm-azure-tpl/releases/download/${HELM_AZURE_TPL_VERSION}/${PLUGIN_DOWNLOAD_FILE}"
PLUGIN_TARGET_PATH="${HELM_PLUGIN_DIR}/${PLUGIN_DOWNLOAD_FILE}"

echo "starting download (via curl)"
echo "  platform: ${HOST_OS}/${HOST_ARCH}"
echo "       url: $PLUGIN_DOWNLOAD_URL"
echo "    target: $PLUGIN_TARGET_PATH"

rm -f -- "$PLUGIN_TARGET_PATH"
curl --fail --location "$PLUGIN_DOWNLOAD_URL" -o "$PLUGIN_TARGET_PATH"
if [ "$?" -ne 0 ]; then
    >&2 echo "[ERROR] failed to download plugin executable"
    exit 1
fi

if [[ ! -f "$PLUGIN_TARGET_PATH" ]]; then
    >&2 echo "[ERROR] installation of executable failed, please report issue"
    exit 1
fi

chmod +x "$PLUGIN_TARGET_PATH"

echo "successfully downloaded executable"
