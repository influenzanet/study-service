#!/usr/bin/env bash
# Standalone memory sampler, useful for sanity checks ("what's the idle
# RSS of the study container?") or for running alongside an export driven
# by a third-party tool you want to test from outside this test environment.
#
# Usage: peak-memory.sh <container-name> [duration-seconds]
set -euo pipefail

container="${1:-lab-study}"
duration="${2:-30}"

echo "Sampling $container memory for ${duration}s (every 100ms)..."

t_end=$(( $(date +%s) + duration ))
peak=0

while [[ $(date +%s) -lt $t_end ]]; do
  v=$(docker exec "$container" sh -c \
    'cat /sys/fs/cgroup/memory.current 2>/dev/null \
     || cat /sys/fs/cgroup/memory/memory.usage_in_bytes 2>/dev/null' 2>/dev/null || echo 0)
  if [[ -n "$v" && "$v" -gt "$peak" ]]; then
    peak=$v
    printf '  peak so far: %s bytes (%.1f MB)\n' \
      "$peak" \
      "$(awk -v b="$peak" 'BEGIN{print b/1024/1024}')"
  fi
  sleep 0.1
done

printf 'final peak: %s bytes (%.1f MB)\n' \
  "$peak" \
  "$(awk -v b="$peak" 'BEGIN{print b/1024/1024}')"
