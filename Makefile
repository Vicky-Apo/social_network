.PHONY: backend-run backend-migrate-up backend-migrate-down backend-migrate-version backend-migrate-goto backend-migrate-force \
	db-up db-down db-ps db-logs db-reset \
	frontend-dev frontend-build frontend-start frontend-lint
BACKEND_DIR := backend
FRONTEND_DIR := frontend
### BACKEND TASKS ###
backend-run:
	$(MAKE) -C $(BACKEND_DIR) run-server

backend-migrate-up:
	$(MAKE) -C $(BACKEND_DIR) migrate-up

backend-migrate-down:
	$(MAKE) -C $(BACKEND_DIR) migrate-down

backend-migrate-version:
	$(MAKE) -C $(BACKEND_DIR) migrate-version

backend-migrate-goto:
	$(MAKE) -C $(BACKEND_DIR) migrate-goto

backend-migrate-force:
	$(MAKE) -C $(BACKEND_DIR) migrate-force


### DATABASE TASKS ###
db-up:
	$(MAKE) -C $(BACKEND_DIR) db-up

db-down:
	$(MAKE) -C $(BACKEND_DIR) db-down

db-ps:
	$(MAKE) -C $(BACKEND_DIR) db-ps

db-logs:
	$(MAKE) -C $(BACKEND_DIR) db-logs

db-reset:
	$(MAKE) -C $(BACKEND_DIR) db-reset

### FRONTEND TASKS ###
frontend-dev:
	npm --prefix $(FRONTEND_DIR) run dev

frontend-build:  
	npm --prefix $(FRONTEND_DIR) run build

frontend-start:
	npm --prefix $(FRONTEND_DIR) run start

frontend-lint:
	npm --prefix $(FRONTEND_DIR) run lint
