#!/bin/bash
exec "$HELM_BIN" azure-tpl apply --stdout "$4"
