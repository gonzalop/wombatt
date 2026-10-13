# Wombatt - Project Instructions

Go application for querying and monitoring solar inverters and battery management systems (BMS).

## Repository & Architecture

- **Entrypoint & Commands**: `main.go` and `cmd/`.
- **Core Logic**: `internal/` (protocols and hardware drivers, e.g., `bms`, `eg4_18kpv`, `solark`).
- **Protocol Documentation**: Reference PDFs are located under `docs/ref/`.
- **Build & Tasks**: Defined in `Makefile` (`make build`, `make test`, `make lint`).

## Workflow & Safety Rules

- Do not modify or touch untracked files not created by you.
- Do not push changes upstream.
- Ensure commit messages do not contain unescaped backticks.

## Important Rules

### Code Navigation & Search Rules (`gograph` & `git grep`)
- Check `gograph stale` before navigating Go symbols. If exit code is non-zero or `.gograph/graph.json` is missing, rebuild with `gograph build .`.
- **NEVER use raw `grep` or `grep -r`**.
- **ALWAYS use `gograph` for Go symbol inspection** — NEVER use `view_file` to read Go symbol implementations or hunt down types/methods:
  - `gograph node <Symbol>`: locate symbol declaration (file + line).
  - `gograph context <Symbol>`: inspect symbol source, callers, callees, and tests without dumping the whole file.
  - `gograph fields <Struct>`: inspect struct fields, types, and tags directly.
  - `gograph callers <Func>`: trace call hierarchy.
- For raw text, strings, or non-Go files, **ALWAYS use `git grep`** (`git grep -n "pattern"`).

### Formatting & Edit Integrity
- Write standard Go code with tab indentation matching `gofmt` conventions so formatting produces 0 diffs.
- Do NOT run bare `make` (or `make all`, `make fmt`, `gofmt -w .`) inside iterative edit-test loops to prevent line-number drift and tool edit failures caused by background file mutations.
- During edit loops, ALWAYS invoke specific targets (`make test`, `make lint`, `make build`) which do not run `fmt`. Run `make fmt` or bare `make` only as a final check right before committing.

### Token Economy & Output Discipline
- Overview first: use `git status -s` and `git diff --stat` before deep inspections.
- Targeted reads: use slice ranges (`StartLine`/`EndLine` in `view_file`) instead of dumping full 500+ line files.
- Surgical edits: use `replace_file_content` with small chunks over full-file overwrites.
- Output brevity: provide concise 1-sentence confirmations after edits; do not re-summarize code changes or print multi-paragraph essays in chat.

