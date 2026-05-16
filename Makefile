.PHONY: all build build-frontend build-backend clean dev-backend dev-frontend

BINARY_NAME = openrgb-batocera-server
GO          = go
NPM         = npm

# ─── Full build ──────────────────────────────────────────────────────────────
all: build

build: build-frontend build-backend

# 1) Build Angular → copy dist into web/
build-frontend:
	@echo ">>> Building Angular frontend..."
	cd frontend && $(NPM) run build
	@echo ">>> Copying to web/..."
	find web/ -mindepth 1 -not -name '.gitkeep' -delete
	cp -r frontend/dist/frontend/browser/. web/

# 2) Build Go binary (embeds web/)
build-backend:
	@echo ">>> Building Go backend..."
	$(GO) build -o $(BINARY_NAME) .
	@echo ">>> Done: ./$(BINARY_NAME)"

# ─── Batocera targets ────────────────────────────────────────────────────────
build-batocera: build-frontend
	@echo ">>> Building for Batocera x86_64 (linux/amd64, static)..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		$(GO) build -ldflags="-s -w" -o $(BINARY_NAME)-batocera .
	@echo ">>> Done: ./$(BINARY_NAME)-batocera  ($$(du -sh $(BINARY_NAME)-batocera | cut -f1))"

build-batocera-arm64: build-frontend
	@echo ">>> Building for Batocera ARM64 (linux/arm64, static)..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		$(GO) build -ldflags="-s -w" -o $(BINARY_NAME)-batocera-arm64 .
	@echo ">>> Done: ./$(BINARY_NAME)-batocera-arm64"

# ─── Development ─────────────────────────────────────────────────────────────

# Run Go backend only (serves whatever is in web/)
dev-backend:
	$(GO) run . --port 8080

# Run Angular dev server with proxy to Go backend on :8080
dev-frontend:
	cd frontend && $(NPM) start

# ─── Clean ───────────────────────────────────────────────────────────────────
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-batocera $(BINARY_NAME)-batocera-arm64
	rm -rf frontend/dist frontend/.angular
	find web/ -mindepth 1 -not -name '.gitkeep' -delete
