#!/bin/bash
set -e
set -o pipefail

>&2 echo ""
>&2 echo "executing azure-tpl for \"$4\":"
./azure-tpl apply --stdout "$4"
EXIT_CODE="$?"
>&2 echo ""
exit "$EXIT_CODE"
