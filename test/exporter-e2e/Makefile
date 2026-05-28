# See README.md for the full workflow.

SHELL := bash

DUMP_DIR ?= ./dump

# Tags used when building the local refactor images.
STUDY_REFACTOR_TAG ?= influweb-local/study-service:refactor
MGMT_REFACTOR_TAG  ?= influweb-local/management-api:refactor

.PHONY: help
help:
	@echo "Targets:"
	@echo "  build-refactor  - docker build refactor images from the working tree"
	@echo "  up-stable       - start the lab with production-tagged images"
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

.PHONY: build-refactor
build-refactor:
	@echo "Building study-service refactor image..."
	docker build --load -t $(STUDY_REFACTOR_TAG) -f ../../build/docker/Dockerfile ../..
	@echo "Building management-api refactor image..."
	docker build --load -t $(MGMT_REFACTOR_TAG)  -f ../../../api-gateway/build/docker/management-api/Dockerfile ../../../api-gateway
	@echo "Done. Use 'make up-refactor' to start the lab."

.PHONY: up-stable
up-stable:
	docker compose up -d
	@echo "Lab up (stable images). Tail logs with 'docker compose logs -f'."

.PHONY: up-refactor
up-refactor:
	STUDY_IMAGE=$(STUDY_REFACTOR_TAG) MGMT_IMAGE=$(MGMT_REFACTOR_TAG) docker compose up -d
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
