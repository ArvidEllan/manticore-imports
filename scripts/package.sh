#!/usr/bin/env bash
set -euo pipefail
./scripts/build.sh
zip -q -j function.zip bootstrap
