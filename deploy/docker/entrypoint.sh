#!/bin/sh
# entrypoint.sh — thin wrapper so `docker run` args are passed straight to superclaw.
#
# Environment variables:
#   ANTHROPIC_API_KEY  — required
#   SUPERCLAW_WORKDIR      — mounted project directory (default: /workspace)
#   SUPERCLAW_CONFIG       — path to superclaw.json inside the container (optional)

set -e

if [ -z "${ANTHROPIC_API_KEY}" ]; then
  echo "error: ANTHROPIC_API_KEY is not set" >&2
  exit 1
fi

WORKDIR="${SUPERCLAW_WORKDIR:-/workspace}"
CONFIG_FLAG=""
if [ -n "${SUPERCLAW_CONFIG}" ]; then
  CONFIG_FLAG="-config ${SUPERCLAW_CONFIG}"
fi

exec superclaw -workdir "${WORKDIR}" ${CONFIG_FLAG} "$@"
