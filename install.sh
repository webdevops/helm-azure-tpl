#!/bin/bash
set -e
set -o pipefail

if [[ -z "$HELM_PLUGIN_DIR" ]]; then
    echo "ERROR: env var HELM_PLUGIN_DIR not set (script must be called by helm)"
    exit 1
fi

FORCE=0
if [[ "$1" == "force" ]]; then
    FORCE=1
fi

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

PLUGIN_DOWNLOAD_FILE="helm-azure-tpl.${HOST_OS}.${HOST_ARCH}${FILE_SUFFIX}"
PLUGIN_DOWNLOAD_URL="https://github.com/webdevops/helm-azure-tpl/releases/download/${HELM_AZURE_TPL_VERSION}/${PLUGIN_DOWNLOAD_FILE}"
PLUGIN_TARGET_PATH="${HELM_PLUGIN_DIR}/helm-azure-tpl${FILE_SUFFIX}"

echo "installing helm-azure-tpl executable"
echo "  platform: ${HOST_OS}/${HOST_ARCH}"
echo "       url: $PLUGIN_DOWNLOAD_URL"
echo "    target: $PLUGIN_TARGET_PATH"

## detect hostedtoolcache
PLUGIN_CACHE_DIR=""
PLUGIN_CACHE_FILE=""
if [[ -n "$RUNNER_TOOL_CACHE" ]]; then
	PLUGIN_CACHE_DIR="${RUNNER_TOOL_CACHE}/helm-azure-tpl/${HELM_AZURE_TPL_VERSION}"
	PLUGIN_CACHE_FILE="${PLUGIN_CACHE_DIR}/${PLUGIN_DOWNLOAD_FILE}"
	echo "    cache: $PLUGIN_CACHE_FILE"
fi

## force (cleanup/upgrade)
if [[ "$FORCE" -eq 1 && -f "$PLUGIN_TARGET_PATH" ]]; then
    echo "removing old executable (update/force mode)"
    rm -f -- "$PLUGIN_TARGET_PATH"
fi

## get from hostedtoolcache
if [[ -n "$PLUGIN_CACHE_FILE" && -e "$PLUGIN_CACHE_FILE" ]]; then
	echo "fetching from RUNNER_TOOL_CACHE ($RUNNER_TOOL_CACHE)"
	cp -- "$PLUGIN_CACHE_FILE" "$PLUGIN_TARGET_PATH"
	chmod +x "$PLUGIN_TARGET_PATH"
fi

## download
if [[ ! -e "$PLUGIN_TARGET_PATH" ]]; then
    echo "starting download (using curl)"
	curl --fail --location "$PLUGIN_DOWNLOAD_URL" -o "$PLUGIN_TARGET_PATH"
	if [ "$?" -ne 0 ]; then
	    >&2 echo "[ERROR] failed to download plugin executable"
	    exit 1
	fi

	## store to hostedtoolcache
	if [[ -n "$PLUGIN_CACHE_FILE" ]]; then
		echo "storing to RUNNER_TOOL_CACHE ($RUNNER_TOOL_CACHE)"
		mkdir -p -- "${PLUGIN_CACHE_DIR}"
		cp -a -- "$PLUGIN_TARGET_PATH" "${PLUGIN_CACHE_FILE}"
	fi
else
    echo "executable already exists, skipping download"
fi

if [[ ! -f "$PLUGIN_TARGET_PATH" ]]; then
    >&2 echo "[ERROR] installation of executable failed, please report issue"
    exit 1
fi

chmod +x "$PLUGIN_TARGET_PATH"

echo "successfully installed executable"
