#!/usr/bin/env bash
# Compare two export runs: file sha + size, elapsed time,
# peak study-container memory.
#
# For CSV files: headers are compared separately, then bodies are sorted before
# sha comparison (old versions of the study service didn't enforce any sort on rows).
# For long-format CSV: only rows with a non-empty value field are compared. The
# old implementation writes placeholder rows (empty value) for questions from other
# survey versions as a side effect of lazy column discovery, the refactor omits
# those semantically empty rows by design.
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
sha_result="$([[ "$a_sha" == "$b_sha" ]] && echo MATCH || echo DIFFER)"
printf '%-26s  %-30s  %-30s  %s\n' 'sha256 (file equality)' "$a_sha" "$b_sha" "$sha_result"

printf '%-26s  %-30s  %-30s  %s\n' 'output size'    "$(fmt_mb "$a_bytes")" "$(fmt_mb "$b_bytes")" "$(delta "$a_bytes" "$b_bytes")"
printf '%-26s  %-30s  %-30s  %s\n' 'wall time (s)'  "$a_secs"             "$b_secs"             "$(delta "$a_secs" "$b_secs")"
printf '%-26s  %-30s  %-30s  %s\n' 'study peak RSS' "$(fmt_mb "$a_peak")"  "$(fmt_mb "$b_peak")"  "$(delta "$a_peak" "$b_peak")"

# --- Content comparison when sha differs ---

if [[ "$sha_result" == "DIFFER" ]]; then
  echo

  if [[ -f "$a/data.csv" && -s "$a/data.csv" && -f "$b/data.csv" && -s "$b/data.csv" ]]; then
    a_header=$(head -1 "$a/data.csv")
    b_header=$(head -1 "$b/data.csv")
    a_ncols=$(echo "$a_header" | tr ',' '\n' | wc -l | tr -d ' ')
    b_ncols=$(echo "$b_header" | tr ',' '\n' | wc -l | tr -d ' ')
    header_result="$([[ "$a_header" == "$b_header" ]] && echo MATCH || echo DIFFER)"
    printf '%-26s  %-30s  %-30s  %s\n' 'csv header' "${a_ncols} cols" "${b_ncols} cols" "$header_result"
    if [[ "$header_result" == "DIFFER" ]]; then
      echo "  headers differ — check column sets separately" >&2
    fi

    eof_err='{"error":"error reading from server: EOF"}'
    a_crashed=0; b_crashed=0
    [[ "$(cat "$a/data.csv")" == "$eof_err" ]] && a_crashed=1
    [[ "$(cat "$b/data.csv")" == "$eof_err" ]] && b_crashed=1

    if (( a_crashed || b_crashed )); then
      (( a_crashed )) && echo "  CRASH: $a/data.csv contains server EOF error — service likely OOM" >&2
      (( b_crashed )) && echo "  CRASH: $b/data.csv contains server EOF error — service likely OOM" >&2
    elif echo "$a_header" | grep -q ',responseSlot,value$'; then
      echo "  (long-format CSV: sorting non-empty-value rows to check content equality...)" >&2
      a_sorted=$(tail -n +2 "$a/data.csv" | grep -v ',$' | sort | shasum -a 256 | awk '{print $1}')
      b_sorted=$(tail -n +2 "$b/data.csv" | grep -v ',$' | sort | shasum -a 256 | awk '{print $1}')
      sorted_label='sha256 (sorted, non-empty)'
      sorted_result="$([[ "$a_sorted" == "$b_sorted" ]] && echo MATCH || echo DIFFER)"
      printf '%-26s  %-30s  %-30s  %s\n' "$sorted_label" "$a_sorted" "$b_sorted" "$sorted_result"
      if [[ "$sorted_result" == "DIFFER" ]]; then
        echo "CSV bodies differ even after sorting."
      fi
    else
      echo "  (sorting CSV bodies to check content equality...)" >&2
      a_sorted=$(tail -n +2 "$a/data.csv" | sort | shasum -a 256 | awk '{print $1}')
      b_sorted=$(tail -n +2 "$b/data.csv" | sort | shasum -a 256 | awk '{print $1}')
      sorted_label='sha256 (sorted body)'
      sorted_result="$([[ "$a_sorted" == "$b_sorted" ]] && echo MATCH || echo DIFFER)"
      printf '%-26s  %-30s  %-30s  %s\n' "$sorted_label" "$a_sorted" "$b_sorted" "$sorted_result"
      if [[ "$sorted_result" == "DIFFER" ]]; then
        echo "CSV bodies differ even after sorting."
      fi
    fi

  else
    echo "Files differ. Quick byte-diff:"
    diff -q "$a"/data.* "$b"/data.* || true
  fi
fi
