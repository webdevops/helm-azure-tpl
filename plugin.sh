#!/bin/bash
set -e
set -o pipefail

println() {
    >&2 echo "$*"
}

println ""
if [[ -n "$GITHUB_ACTION" ]]; then
    println "::group::$4"
fi

println "executing azure-tpl for \"$4\":"
"${HELM_PLUGIN_DIR}/helm-azure-tpl" apply --stdout "$4"
EXIT_CODE="$?"

if [[ -n "$GITHUB_ACTION" ]]; then
    println "::endgroup::"
fi

println ""
exit "$EXIT_CODE"
