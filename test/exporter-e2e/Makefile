# See README.md for the full workflow.

SHELL := bash

DUMP_DIR ?= ./dump

# Host architecture (arm64/aarch64 -> arm64, x86_64/amd64 -> amd64). The lab
# builds all Go services natively for this arch; see scripts/build-images.sh.
ARCH := $(shell uname -m | sed -e 's/aarch64/arm64/' -e 's/x86_64/amd64/')

# Arch-suffixed local image tags. Must match scripts/build-images.sh. What each
# tag is built from is declared per-service in images.conf.
STUDY_STABLE_IMAGE      ?= influweb-local/study-service:stable-$(ARCH)
MGMT_STABLE_IMAGE       ?= influweb-local/management-api:stable-$(ARCH)
USERMGMT_STABLE_IMAGE   ?= influweb-local/user-management:stable-$(ARCH)
STUDY_REFACTOR_IMAGE    ?= influweb-local/study-service:refactor-$(ARCH)
MGMT_REFACTOR_IMAGE     ?= influweb-local/management-api:refactor-$(ARCH)
USERMGMT_REFACTOR_IMAGE ?= influweb-local/user-management:refactor-$(ARCH)

.PHONY: help
help:
	@echo "Host architecture: $(ARCH)"
	@echo "Targets:"
	@echo "  build-stable    - build the stable image set (sources declared in images.conf)"
	@echo "  build-refactor  - build the refactor image set (sources declared in images.conf)"
	@echo "  build-all       - build both stable and refactor image sets"
	@echo "  up-stable       - start the lab with locally-built stable images"
	@echo "  up-refactor     - start the lab with locally-built refactor images"
	@echo "  down            - stop the lab (data preserved)"
	@echo "  nuke            - stop the lab AND drop the mongo volume"
	@echo "  restore         - mongorestore the contents of \$$DUMP_DIR into the lab mongo"
	@echo "  login           - obtain an admin JWT and save it to .token (.env credentials)"
	@echo "  export-wide     - download wide CSV; saves to out/<run>/wide.csv and times+peaks it"
	@echo "  export-long     - download long-format CSV"
	@echo "  export-json     - download flat JSON"
	@echo "  export-all      - run all three exports"
	@echo "  peak-stable     - sample memory of the study container while idle (sanity)"
	@echo "  diff            - sha256 + size compare two runs (RUN_A=stable RUN_B=refactor)"
	@echo "  ps              - show lab container status + memory usage"

.PHONY: build-stable
build-stable:
	./scripts/build-images.sh stable

.PHONY: build-refactor
build-refactor:
	./scripts/build-images.sh refactor

.PHONY: build-all
build-all:
	./scripts/build-images.sh all

.PHONY: up-stable
up-stable:
	./scripts/build-images.sh stable
	STUDY_IMAGE=$(STUDY_STABLE_IMAGE) MGMT_IMAGE=$(MGMT_STABLE_IMAGE) USERMGMT_IMAGE=$(USERMGMT_STABLE_IMAGE) docker compose up -d
	@echo "Lab up (stable images). Tail logs with 'docker compose logs -f'."

.PHONY: up-refactor
up-refactor:
	./scripts/build-images.sh refactor
	STUDY_IMAGE=$(STUDY_REFACTOR_IMAGE) MGMT_IMAGE=$(MGMT_REFACTOR_IMAGE) USERMGMT_IMAGE=$(USERMGMT_REFACTOR_IMAGE) docker compose up -d
	@echo "Lab up (refactor images). Tail logs with 'docker compose logs -f'."

.PHONY: down
down:
	docker compose down

.PHONY: nuke
nuke:
	docker compose down -v
	rm -rf out .token

.PHONY: ps
ps:
	@docker compose ps
	@echo
	@docker stats --no-stream --format 'table {{.Name}}\t{{.MemUsage}}\t{{.CPUPerc}}' \
	  lab-mongo lab-user-management lab-study lab-management-api 2>/dev/null || true

.PHONY: restore
restore:
	@test -d $(DUMP_DIR) || (echo "ERROR: $(DUMP_DIR) not found — see README §2"; exit 1)
	@test -n "$$(ls -A $(DUMP_DIR) 2>/dev/null)" || (echo "ERROR: $(DUMP_DIR) is empty"; exit 1)
	@echo "Restoring dump from $(DUMP_DIR) into lab mongo..."
	docker compose exec -T mongo bash -c 'mongorestore --username "$$MONGO_INITDB_ROOT_USERNAME" --password "$$MONGO_INITDB_ROOT_PASSWORD" --authenticationDatabase admin --drop /dump'
	@echo "Restore complete. Run 'make ps' to inspect."

.PHONY: login
login:
	./scripts/login.sh

.PHONY: export-wide
export-wide:
	./scripts/run-export.sh wide

.PHONY: export-long
export-long:
	./scripts/run-export.sh long

.PHONY: export-json
export-json:
	./scripts/run-export.sh json

.PHONY: export-all
export-all: export-wide export-long export-json

.PHONY: peak-stable
peak-stable:
	./scripts/peak-memory.sh lab-study 30

.PHONY: diff
diff:
	@test -n "$(RUN_A)" -a -n "$(RUN_B)" || (echo "Usage: make diff RUN_A=<dir> RUN_B=<dir>"; exit 1)
	./scripts/diff-runs.sh $(RUN_A) $(RUN_B)
