SHOTS_DIR ?= screenshots
SHOT_W ?= 1920
SHOT_H ?= 1080
SHOT_TICKS ?= 60
SHOT_SEED ?= 1
CLICK_COUNT ?= 40
CLICK_SEED ?= 1234567890

# Committed golden frames, used to spot unintended visual changes (a dependency
# bump, a render regression). Deliberately smaller than the `screenshots`
# defaults: these live in git forever, and 800x600 still catches anything worth
# catching. See testdata/screenshots/README.md.
BASELINE_DIR ?= testdata/screenshots
BASELINE_W ?= 800
BASELINE_H ?= 600

EXAMPLE_DIRS := $(wildcard examples/*) $(wildcard visual_tests/*)

# Packages whose visual output only appears after mouse input; these get CLICK_COUNT
# scripted left-click press/release cycles at random (but seeded, deterministic) points.
MOUSE_DRIVEN_DIRS := visual_tests/kdtree_mouse visual_tests/quadtree_mouse

# Packages whose frame differs run to run at identical settings, so byte
# comparison says nothing. Their baselines are still committed as a visual
# reference, but check-baselines skips them. reaction_diffusion and
# shader_photo are GPU shader passes; nearest_neighbor animates off wall-clock
# time.
NONDETERMINISTIC_DIRS := examples/reaction_diffusion examples/shader_photo visual_tests/nearest_neighbor

# render_shots <out-dir> <width> <height>
define render_shots
@mkdir -p $(1)
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
		-w $(2) \
		-h $(3) \
		-out $(1)/$$name.png \
	|| echo "    FAILED: $$d"; \
done
endef

.PHONY: screenshots
screenshots:
	$(call render_shots,$(SHOTS_DIR),$(SHOT_W),$(SHOT_H))

.PHONY: clean-screenshots
clean-screenshots:
	rm -rf $(SHOTS_DIR)

# Regenerate the committed golden frames. Review the diff before committing:
# every changed pixel is either a fix or a regression, and only you can say
# which.
.PHONY: baselines
baselines:
	$(call render_shots,$(BASELINE_DIR),$(BASELINE_W),$(BASELINE_H))

# Re-render into a scratch directory and byte-compare against the committed
# baselines, skipping NONDETERMINISTIC_DIRS. Exits non-zero on any difference.
.PHONY: check-baselines
check-baselines:
	@tmp=$$(mktemp -d) && trap 'rm -rf "$$tmp"' EXIT && \
	$(MAKE) --no-print-directory baselines BASELINE_DIR="$$tmp" >/dev/null 2>&1 && \
	skip=""; for m in $(NONDETERMINISTIC_DIRS); do skip="$$skip $$(echo $$m | tr '/' '_')"; done; \
	rc=0; \
	for f in $(BASELINE_DIR)/*.png; do \
		name=$$(basename $$f .png); \
		case " $$skip " in *" $$name "*) echo "skip     $$name"; continue;; esac; \
		if [ ! -f "$$tmp/$$name.png" ]; then echo "MISSING  $$name (render failed)"; rc=1; \
		elif cmp -s "$$f" "$$tmp/$$name.png"; then echo "ok       $$name"; \
		else echo "CHANGED  $$name"; rc=1; fi; \
	done; \
	[ $$rc -eq 0 ] && echo "baselines match" || echo "baselines differ; run 'make baselines' and review"; \
	exit $$rc

.PHONY: lint
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found; install it: https://golangci-lint.run/welcome/install/"; \
		exit 1; \
	}
	golangci-lint run
