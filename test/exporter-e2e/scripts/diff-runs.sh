#!/usr/bin/env bash
# Compare two export runs: file sha + size, elapsed time,
# peak study-container memory. Use to confirm the refactor produces
# byte-identical output AND lower memory than stable.
#
# Usage: diff-runs.sh <run-dir-a> <run-dir-b>
set -euo pipefail

a="${1:?usage: diff-runs.sh <run-dir-a> <run-dir-b>}"
b="${2:?usage: diff-runs.sh <run-dir-a> <run-dir-b>}"

for d in "$a" "$b"; do
  [[ -d "$d" && -f "$d/meta.txt" ]] || { echo "missing $d/meta.txt"; exit 1; }
done

get() { awk -v k="$1:" '$1==k{$1=""; sub(/^ /,""); print; exit}' "$2"; }

a_sha=$(get sha256 "$a/meta.txt")
b_sha=$(get sha256 "$b/meta.txt")
a_bytes=$(get bytes "$a/meta.txt" | awk '{print $1}')
b_bytes=$(get bytes "$b/meta.txt" | awk '{print $1}')
a_secs=$(get elapsed_seconds "$a/meta.txt")
b_secs=$(get elapsed_seconds "$b/meta.txt")
a_peak=$(get study_peak_memory_bytes "$a/meta.txt" | awk '{print $1}')
b_peak=$(get study_peak_memory_bytes "$b/meta.txt" | awk '{print $1}')

fmt_mb() { awk -v b="$1" 'BEGIN{printf "%.1f MB", b/1024/1024}'; }
delta()  { awk -v a="$1" -v b="$2" 'BEGIN{ if(a==0){print "—"} else {printf "%+.1f%%", (b-a)/a*100} }'; }

printf '%-26s  %-30s  %-30s  %s\n' field "$(basename "$a")" "$(basename "$b")" delta
printf '%-26s  %-30s  %-30s  %s\n' '----' '----' '----' '----'
printf '%-26s  %-30s  %-30s  %s\n' 'sha256 (file equality)' "$a_sha" "$b_sha" \
  "$([[ "$a_sha" == "$b_sha" ]] && echo MATCH || echo DIFFER)"
printf '%-26s  %-30s  %-30s  %s\n' 'output size'      "$(fmt_mb "$a_bytes")" "$(fmt_mb "$b_bytes")" "$(delta "$a_bytes" "$b_bytes")"
printf '%-26s  %-30s  %-30s  %s\n' 'wall time (s)'    "$a_secs"             "$b_secs"             "$(delta "$a_secs" "$b_secs")"
printf '%-26s  %-30s  %-30s  %s\n' 'study peak RSS'   "$(fmt_mb "$a_peak")"  "$(fmt_mb "$b_peak")"  "$(delta "$a_peak" "$b_peak")"

if [[ "$a_sha" != "$b_sha" ]]; then
  echo
  echo "Files differ. Quick byte-diff:"
  diff -q "$a"/data.* "$b"/data.* || true
fi
