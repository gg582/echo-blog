.PHONY: build test install update uninstall clean

FRONTEND_DIR := blog-frontend
BACKEND_DIR  := blog-backend
BINARY_NAME  := echo-blog
SERVICE      := echo-blog
SYSTEMD_DIR  := /etc/systemd/system
DATA_DIR     := data

build:
	@echo "==> Building frontend..."
	cd $(FRONTEND_DIR) && npm ci && npm run build
	@echo "==> Building Go backend..."
	cd $(BACKEND_DIR) && CGO_ENABLED=1 go build -o $(BINARY_NAME) .
	@echo "Build complete: $(BACKEND_DIR)/$(BINARY_NAME)"

test:
	cd $(BACKEND_DIR) && go vet ./... && go test ./...

# install expects this checkout to live at /opt/echo-blog (see deploy/echo-blog.service).
# Content is kept in $(DATA_DIR)/, which is seeded from the repository only when
# it does not exist yet, so reinstalling never overwrites posts or accounts.
install: build
	@if [ $$(id -u) -ne 0 ]; then \
		echo "Error: 'make install' must be run as root (try: sudo make install)"; \
		exit 1; \
	fi
	@if [ ! -d $(DATA_DIR) ]; then \
		echo "==> Seeding $(DATA_DIR)/ from the repository..."; \
		mkdir -p $(DATA_DIR); \
		cp -a $(BACKEND_DIR)/posts $(BACKEND_DIR)/about $(BACKEND_DIR)/contact $(DATA_DIR)/; \
		cp -a $(BACKEND_DIR)/auth.db $(DATA_DIR)/ 2>/dev/null || true; \
	else \
		echo "==> Keeping existing $(DATA_DIR)/"; \
	fi
	@echo "==> Installing systemd service..."
	install -m 644 deploy/echo-blog.service $(SYSTEMD_DIR)/$(SERVICE).service
	systemctl daemon-reload
	systemctl enable $(SERVICE)
	systemctl restart $(SERVICE)
	@echo ""
	@echo "Installation complete."
	@echo "  - Service : $(SERVICE)"
	@echo "  - Secrets : /etc/echo-blog/echo-blog.env (AUTH_SECRET=...)"
	@echo "  - Logs    : /opt/echo-blog/server.log, /opt/echo-blog/error.log"
	@echo "  - Update  : sudo make update"

update:
	./deploy/update.sh

# uninstall removes the service only; $(DATA_DIR)/ is left in place.
uninstall:
	@if [ $$(id -u) -ne 0 ]; then \
		echo "Error: 'make uninstall' must be run as root (try: sudo make uninstall)"; \
		exit 1; \
	fi
	systemctl stop $(SERVICE) 2>/dev/null || true
	systemctl disable $(SERVICE) 2>/dev/null || true
	rm -f $(SYSTEMD_DIR)/$(SERVICE).service
	systemctl daemon-reload
	@echo "Service removed. Content in $(DATA_DIR)/ was kept."

clean:
	@echo "==> Cleaning build artifacts..."
	cd $(FRONTEND_DIR) && rm -rf build
	cd $(BACKEND_DIR) && rm -f $(BINARY_NAME)
	@echo "Clean complete."
