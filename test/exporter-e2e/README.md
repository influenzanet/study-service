# Exporter end-to-end tests

Spins up a minimal, self-contained Influenzanet stack on the local machine in order to perform e2e testing of the exporter. 
The export can be driven by the ifn-cli or from the provided script (`run-export.sh`) against `http://localhost:3232`, with the option of swapping between the stable production images and the locally-built refactor images via a single environment variable.

Memory of each container is captured throughout the export, so it's possible compare peak RSS between stable and refactor runs and confirm the output bytes are identical.

## What's in this directory

```
local-lab/
├── docker-compose.yml      # 4 services: mongo, user-management, study, management-api
├── .env.example            # copy to .env and edit
├── Makefile                # workflow entrypoints (build, up, login, export, diff)
├── scripts/
│   ├── login.sh            # POST /v1/auth/login-with-email, save .token
│   ├── run-export.sh       # drive one export, capture file + memory log
│   ├── peak-memory.sh      # standalone cgroup memory sampler
│   └── diff-runs.sh        # head-to-head comparison of two run dirs
├── dump/                   # (you create) holds the mongodump output
└── out/                    # (created on first export) one subdir per run
```

---

## Prerequisites

- Docker or Podman with `docker compose` v2 (or `podman compose`).
- `jq` and `curl`.

---

## One-time setup

### 1. Configure the environment

```bash
cd study-service/tes/exporter-e2e
cp .env.example .env
$EDITOR .env
```

Fill in:

- `JWT_TOKEN_KEY`: any random string, do not use a key used in production.
- `STUDY_GLOBAL_SECRET`: same
- `DB_NAME_PREFIX`: leave empty unless you are using a dump for which the prefix was set.
- `ADMIN_EMAIL`, `ADMIN_PASSWORD`: credentials of an account that already exists in the dumped DB and has researcher/owner role on the study.
- `INSTANCE_ID`: Must match the one used in the dumbped DB.
- `STUDY_KEY` / `SURVEY_KEY`: the export target.
- `FROM` / `UNTIL`: optional unix-second filters; leave blank for all-time.

### 2. Dump MongoDB

You can either create syntethic data or dump an existing Mongo DB.

From a host that can reach the cluster's Mongo:

```bash
# Example: dump the user DB + the study DB for one instance.
mongodump \
  --uri "mongodb://USER:PASS@mongo-atlas-service:27017" \
  --db "${DB_NAME_PREFIX}${INSTANCE_ID}_users" \
  --out /tmp/influweb-dump \
  --authenticationDatabase=admin

mongodump \
  --uri "mongodb://USER:PASS@mongo-atlas-service:27017" \
  --db "${DB_NAME_PREFIX}${INSTANCE_ID}_studyDB" \
  --out /tmp/influweb-dump \
  --authenticationDatabase=admin

# tar it for transfer
tar -czf dump.tar.gz -C /tmp dump
```

The minimum DBs the lab needs are:

- `${DB_NAME_PREFIX}${INSTANCE_ID}_users`: for login and JWT issuance
- `${DB_NAME_PREFIX}${INSTANCE_ID}_studyDB`: for the survey definitions, study permissions, and the responses being exported
- `${DB_NAME_PREFIX}global-infos`: contains instance metadata. 

Copy the inner contents of the dump into `./dump/`. The layout must be `dump/<dbname>/<collection>.bson`, which is what `mongorestore /dump` will read.

```bash
mkdir -p study-service/test/exporter-e2e/dump
tar -xzf dump.tar.gz -C /tmp
cp -r /tmp/dump/. study-service/test/exporter-e2e/dump/
```

**IMPORTANT** disable 2FA for the user

> **Privacy note**: If you are using a real database remember to meet whatever data-handling policy applies. Add `dump/` and `out/` to your local `.gitignore` and make sure to never commit them

### 3. Build the refactor images from the working tree

```bash
make build-refactor
```

This produces two locally-tagged images:

- `influweb-local/study-service:refactor`
- `influweb-local/management-api:refactor`

Override the tags via `STUDY_REFACTOR_TAG` / `MGMT_REFACTOR_TAG` if needed.

### 4. First run with stable images, then restore the dump

```bash
make up-stable    # pulls and starts the 4 containers (production-tagged images)
make ps           # sanity: all four should be Up
make restore      # mongorestore /dump → lab mongo
```

### 5. Login

```bash
make login        # writes .token
```

If login fails:
- **`account not confirmed`** — the prod admin account has `accountConfirmedAt = null` in the dump. Either set it manually:
  ```bash
  docker compose exec mongo mongosh --eval '
    db = db.getSiblingDB("default_users");
    db.users.updateOne(
      {"account.accountId":"YOUR_EMAIL"},
      {$set:{"account.accountConfirmedAt": NumberLong(Math.floor(Date.now()/1000))}}
    );
  '
  ```
- **`invalid credentials`** — password mismatch. The Argon2 hash in the dump is independent of our local JWT key, so the actual production password must be used.
- **`no users found`** — the instance ID or DB prefix doesn't match. Check `<instance_id>users` exists in mongo: `docker compose exec mongo mongosh --eval 'db.adminCommand("listDatabases")'`.

---

## Running an end to end test

### Single run

```bash
# Run the export
RUN_LABEL=stable make export-wide
RUN_LABEL=stable make export-long
RUN_LABEL=stable make export-json
```

Each run creates `out/<timestamp>-stable-<format>/` with:

- `data.csv` (or `.json`):
- `memory.log`: one line per 100 ms with `<timestamp> <bytes_in_use>`. Used to compute peak RSS.
- `meta.txt`: http_code, elapsed_seconds, bytes, sha256, peak memory, and which container image was running.

### Head-to-head comparison: stable vs refactor

```bash
# stable
make up-stable
make login            # token issued by the running user-management, saved to .token
RUN_LABEL=stable make export-wide
RUN_LABEL=stable make export-long
RUN_LABEL=stable make export-json

# Switch images (keeps mongo + the existing dump in place)
make up-refactor      # restarts study + management-api with refactor images,
                      # leaves mongo + user-management running

make login

RUN_LABEL=refactor make export-wide
RUN_LABEL=refactor make export-long
RUN_LABEL=refactor make export-json

# Compare:
make diff RUN_A=out/<stable-wide-dir> RUN_B=out/<refactor-wide-dir>
make diff RUN_A=out/<stable-long-dir> RUN_B=out/<refactor-long-dir>
make diff RUN_A=out/<stable-json-dir> RUN_B=out/<refactor-json-dir>
```

`make diff` prints sha256, file size, wall time, and peak study-container RSS for both runs with percent deltas. If the sha256s match, the refactor is byte-identical.

### Exporting from ifn-cli

The container's HTTP surface is `http://localhost:3232`. Configure ifn-cli accordingly

While the request is in flight, in another terminal:

```bash
./scripts/peak-memory.sh lab-study 120
./scripts/peak-memory.sh lab-management-api 120   # also worth checking the gateway
```

The gateway's peak RSS is the part that confirms the streaming flush actually works end-to-end, pre-refactor it climbed to ~the file size while post-refactor it should stay flat at a few tens of MB regardless of file size.

---

## Memory analysis: what to look at

For each format, compare across stable vs refactor:

| metric | meaning | expected change |
|---|---|---|
| `study_peak_memory_bytes` in `meta.txt` | cgroup `memory.current` peak of `lab-study` | wide CSV: small drop; **long CSV: large drop**; JSON: medium drop |
| `lab-management-api` peak (via `peak-memory.sh`) | streamed vs buffered on the gateway | **flat (~30-50 MB) on refactor; climbs to ~file size on stable** |
| `elapsed_seconds` | wall time to fully download | mostly flat, possibly slight improvement |
| `sha256` | byte-identity check | **must match** between stable and refactor |

For a really sharp view of the memory curve, plot `memory.log` columns 1 (time) vs 2 (bytes). Stable will show a ramp while refactor will plateau much earlier (or stay flat for the gateway).

---

## Tearing down

```bash
make down       # stops containers, keeps the mongo volume and dump
make nuke       # stops AND wipes the mongo volume, deletes out/ and .token
                # use this when you're done with the lab or want a clean reimport
```

---

## Troubleshooting

- **`docker compose up` fails with image not found for `influweb-local/...:refactor`**: you didn't run `make build-refactor` yet. Or the tag in `.env` doesn't match the tag you built.
- **management-api logs a `dial tcp unused:5004: ... no such host` warning**: harmless. messaging-service isn't included in the lab and management-api tolerates it being absent for export flows.
- **`account not confirmed` on login**: see step 5 above.
- **`permission denied` on a protected route**: the admin account must have `STUDY_ROLE_OWNER` or `STUDY_ROLE_MAINTAINER` (or be a USER_ROLE_ADMIN). Check `db.studies.findOne({key: "<STUDY_KEY>"})` in the lab mongo.
- **Mongo container OOMs on a big restore**: bump `mongo.deploy.resources.limits.memory` in `docker-compose.yml`.
- **You changed code and `make up-refactor` still uses the old image**: rebuild — `make build-refactor` first.
- **You want to see the actual gRPC chunk traffic between management-api and study**: `docker compose logs -f study management-api` shows both sides' stdout. Inside the study container, set `LOG_LEVEL=debug` for verbose output.



