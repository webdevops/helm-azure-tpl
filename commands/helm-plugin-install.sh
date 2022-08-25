#!/bin/sh
set -e

HELM_AZURE_TPL_VERSION=$(sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' "${HELM_PLUGIN_DIR}/plugin.yaml")

HOST_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
HOST_ARCH=$(uname -m | tr '[:upper:]' '[:lower:]')
PLUGIN_DOWNLOAD_FILE="helm-azure-tpl.${HOST_OS}.${HOST_ARCH}"
PLUGIN_DOWNLOAD_URL="https://github.com/webdevops/helm-azure-tpl/releases/download/${HELM_AZURE_TPL_VERSION}/${PLUGIN_DOWNLOAD_FILE}"
PLUGIN_TARGET_PATH="${HELM_PLUGIN_DIR}/${PLUGIN_DOWNLOAD_FILE}"

echo "download from: $PLUGIN_DOWNLOAD_URL"
echo "target path: $PLUGIN_TARGET_PATH"

curl --location "$PLUGIN_DOWNLOAD_URL" -o "$PLUGIN_TARGET_PATH"
chmod +x "$PLUGIN_TARGET_PATH"
