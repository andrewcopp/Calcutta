#!/bin/bash

set -e

# Wrapper around backend/scripts/run_sandbox.sh for the common end-to-end workflow.
# Defaults:
# - mode=report
# - exclude-entry-name="Andrew Copp"
# - train-years=0 (all history)
# - output to /tmp (macOS: /private/tmp)

OUT_DEFAULT="/tmp/calcutta_sandbox_report.md"

OUT_PATH=""
ARGS=("$@")

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# If the caller provided -out, respect it.
for ((i=0; i<${#ARGS[@]}; i++)); do
  if [ "${ARGS[$i]}" = "-out" ]; then
    j=$((i+1))
    if [ $j -lt ${#ARGS[@]} ]; then
      OUT_PATH="${ARGS[$j]}"
    fi
  fi
  if [[ "${ARGS[$i]}" == -out=* ]]; then
    OUT_PATH="${ARGS[$i]#-out=}";
  fi
done

if [ -z "${OUT_PATH}" ]; then
  OUT_PATH="${OUT_DEFAULT}"
  ARGS=("-out" "${OUT_PATH}" "${ARGS[@]}")
fi

"${SCRIPT_DIR}/run_sandbox.sh" \
  -mode report \
  -pred-model kenpom \
  -sigma 11.0 \
  -exclude-entry-name "Andrew Copp" \
  -train-years 0 \
  "${ARGS[@]}"

echo "Wrote report to: ${OUT_PATH}"
