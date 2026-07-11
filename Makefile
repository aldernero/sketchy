SHOTS_DIR ?= screenshots
SHOT_W ?= 1920
SHOT_H ?= 1080
SHOT_TICKS ?= 60
SHOT_SEED ?= 1
CLICK_COUNT ?= 40
CLICK_SEED ?= 1234567890

EXAMPLE_DIRS := $(wildcard examples/*) $(wildcard visual_tests/*)

# Packages whose visual output only appears after mouse input; these get CLICK_COUNT
# scripted left-click press/release cycles at random (but seeded, deterministic) points.
MOUSE_DRIVEN_DIRS := visual_tests/kdtree_mouse visual_tests/quadtree_mouse

.PHONY: screenshots
screenshots:
	@mkdir -p $(SHOTS_DIR)
	@for d in $(EXAMPLE_DIRS); do \
		[ -d "$$d" ] || continue; \
		name=$$(echo $$d | tr '/' '_'); \
		clicks=0; \
		for m in $(MOUSE_DRIVEN_DIRS); do \
			[ "$$d" = "$$m" ] && clicks=$(CLICK_COUNT); \
		done; \
		echo "==> $$d (clicks=$$clicks)"; \
		go run ./tools/vshot \
			-pkg ./$$d \
			-ticks $(SHOT_TICKS) \
			-seed $(SHOT_SEED) \
			-clicks $$clicks \
			-click-seed $(CLICK_SEED) \
			-w $(SHOT_W) \
			-h $(SHOT_H) \
			-out $(SHOTS_DIR)/$$name.png \
		|| echo "    FAILED: $$d"; \
	done

.PHONY: clean-screenshots
clean-screenshots:
	rm -rf $(SHOTS_DIR)

.PHONY: lint
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found; install it: https://golangci-lint.run/welcome/install/"; \
		exit 1; \
	}
	golangci-lint run
