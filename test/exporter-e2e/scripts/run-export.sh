#!/usr/bin/env bash
# Drive one export through the real HTTP path while simultaneously capturing
# the study and management-api containers' peak resident memory and wall time.
#
# Usage: run-export.sh {wide|long|json}
#
# Produces under out/<timestamp>-<format>/:
#   - the downloaded CSV/JSON file
#   - memory.log  (one docker-stats RSS reading per ~500ms for both the study
#                  and management-api containers, with peaks printed)
#   - meta.txt    (timing, byte count, sha256, image tags in use)
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ -f .env ]]; then
  # shellcheck disable=SC1091
  set -a; source .env; set +a
fi

format="${1:?usage: run-export.sh wide|long|json}"

case "$format" in
  wide) path="response";              ext="csv";  ctype="text/csv" ;;
  long) path="response/long-format";  ext="csv";  ctype="text/csv" ;;
  json) path="response/json";         ext="json"; ctype="application/json" ;;
  *) echo "unknown format: $format" >&2; exit 1 ;;
esac

: "${STUDY_KEY:?STUDY_KEY not set in .env}"
: "${SURVEY_KEY:?SURVEY_KEY not set in .env}"

if [[ ! -f .token ]]; then
  echo "No .token found — run 'make login' first." >&2
  exit 1
fi
token=$(cat .token)

BASE_URL="${BASE_URL:-http://localhost:3232}"

# Build URL with optional filters.
#   FROM / UNTIL — unix-second bounds on submittedAt
#   SEPARATOR    — gateway query "sep" (default at gateway: "-"). Use "." to
#                  get pandas-friendly column names like "intake.Q1.0".
#   SHORT_KEYS   — gateway query "shortKeys" (default at gateway: "true"). Set
#                  to "false" to keep the survey-key prefix in column names
#                  (e.g. "intake.Q1" instead of bare "Q1").
url="${BASE_URL}/v1/data/${STUDY_KEY}/survey/${SURVEY_KEY}/${path}"
q=()
[[ -n "${FROM:-}"        ]] && q+=("from=${FROM}")
[[ -n "${UNTIL:-}"       ]] && q+=("until=${UNTIL}")
[[ -n "${SEPARATOR:-}"   ]] && q+=("sep=${SEPARATOR}")
[[ -n "${SHORT_KEYS:-}"  ]] && q+=("shortKeys=${SHORT_KEYS}")
[[ "${#q[@]}" -gt 0 ]] && url="${url}?$(IFS='&'; echo "${q[*]}")"

# Output directory tagged by run-label (stable/refactor) if provided.
label="${RUN_LABEL:-run}"
run_dir="out/$(date +%Y%m%d-%H%M%S)-${label}-${format}"
mkdir -p "$run_dir"

# Memory sampler: poll the study and management-api containers via `docker stats`
# every 500 ms. Using docker stats rather than cgroup files inside the container
# so it works on macOS (Docker Desktop / Podman) where cgroup paths are not
# exposed. Each line: "<epoch> <study_bytes> <mgmt_bytes>".
sample_pid=""
start_memory_sampler() {
  (
    while true; do
      docker stats --no-stream --format '{{.Name}} {{.MemUsage}}' \
        lab-study lab-management-api 2>/dev/null \
      | awk -v ts="$(date +%s.%N)" '
          function tobytes(s,   val, unit, mult) {
            val = s; gsub(/[^0-9.]/, "", val)
            unit = s; gsub(/[0-9.]/,  "", unit)
            if      (unit == "B")   mult = 1
            else if (unit == "kB")  mult = 1000
            else if (unit == "MB")  mult = 1000000
            else if (unit == "GB")  mult = 1000000000
            else if (unit == "KiB") mult = 1024
            else if (unit == "MiB") mult = 1048576
            else if (unit == "GiB") mult = 1073741824
            else if (unit == "TiB") mult = 1099511627776
            else                    mult = 1
            return int(val * mult)
          }
          $1 == "lab-study"          { study = tobytes($2) }
          $1 == "lab-management-api" { mgmt  = tobytes($2) }
          END { printf "%s %d %d\n", ts, study+0, mgmt+0 }
        '
      sleep 0.5 || true
    done
  ) > "$run_dir/memory.log" &
  sample_pid=$!
}

stop_memory_sampler() {
  if [[ -n "$sample_pid" ]] && kill -0 "$sample_pid" 2>/dev/null; then
    kill "$sample_pid" 2>/dev/null || true
    wait "$sample_pid" 2>/dev/null || true
  fi
}
trap stop_memory_sampler EXIT

# Record which image is actually running for traceability.
study_image=$(docker inspect -f '{{.Config.Image}}' lab-study 2>/dev/null || echo unknown)
mgmt_image=$(docker inspect -f '{{.Config.Image}}' lab-management-api 2>/dev/null || echo unknown)

echo "GET $url"
echo "study image:          $study_image"
echo "management-api image: $mgmt_image"

start_memory_sampler

t0=$(date +%s.%N)
http_code=$(curl -sS -w '%{http_code}' \
  -H "Authorization: Bearer $token" \
  -H "Accept: $ctype" \
  -o "$run_dir/data.${ext}" \
  "$url") || http_code="000"
t1=$(date +%s.%N)

stop_memory_sampler
trap - EXIT

elapsed=$(awk -v a="$t0" -v b="$t1" 'BEGIN{printf "%.3f", b-a}')
bytes=$(stat -f%z "$run_dir/data.${ext}" 2>/dev/null || stat -c%s "$run_dir/data.${ext}" 2>/dev/null || echo 0)
sha=$(shasum -a 256 "$run_dir/data.${ext}" | awk '{print $1}')

if [[ -s "$run_dir/memory.log" ]]; then
  study_peak=$(awk 'NF==3 && $2+0>0 {if($2+0>m){m=$2+0}} END{printf "%d", m}' "$run_dir/memory.log")
  mgmt_peak=$(awk 'NF==3 && $3+0>0 {if($3+0>m){m=$3+0}} END{printf "%d", m}' "$run_dir/memory.log")
  samples=$(wc -l <"$run_dir/memory.log" | awk '{print $1}')
else
  study_peak=0
  mgmt_peak=0
  samples=0
fi

# Pretty bytes (MB).
study_peak_mb=$(awk -v b="$study_peak" 'BEGIN{printf "%.1f", b/1024/1024}')
mgmt_peak_mb=$(awk -v b="$mgmt_peak" 'BEGIN{printf "%.1f", b/1024/1024}')
file_mb=$(awk -v b="$bytes" 'BEGIN{printf "%.1f", b/1024/1024}')

{
  echo "format:                 $format"
  echo "url:                    $url"
  echo "http_code:              $http_code"
  echo "elapsed_seconds:        $elapsed"
  echo "bytes:                  $bytes  (${file_mb} MB)"
  echo "sha256:                 $sha"
  echo "study_peak_memory_bytes: $study_peak  (${study_peak_mb} MB)"
  echo "management_api_peak_memory_bytes: $mgmt_peak  (${mgmt_peak_mb} MB)"
  echo "memory_samples:         $samples"
  echo "study_image:            $study_image"
  echo "management_api_image:   $mgmt_image"
} | tee "$run_dir/meta.txt"

echo
echo "Saved to: $run_dir"

# First lines of the CSV/JSON for a sanity-check that the file is well-formed.
if [[ "$ext" == "csv" ]]; then
  echo
  echo "--- first 3 lines ---"
  head -n 3 "$run_dir/data.${ext}"
fi
