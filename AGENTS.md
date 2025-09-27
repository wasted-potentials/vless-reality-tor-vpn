# Repository Guidelines

## Project Structure & Module Organization
- `cmd/vpnd` holds the long-running daemon that manages users, renders Xray configs, and supervises the external `xray` process.
- `cmd/keygen` provides a minimal X25519 helper for environments where `xray x25519` is unavailable.
- `internal/api`, `internal/users`, and `internal/xray` slice responsibilities between HTTP handlers, JSON-backed user storage, and process/config integration logic.
- `deployments/docker` and `deployments/systemd` contain production manifests; `deployments/tor` ships a sample `torrc`.
- `example/config.yaml` is the canonical starting point for settings, while `scripts/vless_url.sh` prints a ready-to-share VLESS URL.

## Build, Test, and Development Commands
- `make build` compiles static binaries into `./bin` for `vpnd` and `keygen` (CGO disabled for portability).
- `make run` runs `vpnd` against `example/config.yaml`; tweak the path when iterating on custom configs.
- `make docker` builds the container image using `deployments/docker/Dockerfile`.
- `go test ./...` executes Go unit tests; add targeted packages (for example `go test ./internal/users`).

## Coding Style & Naming Conventions
- Format Go code with `gofmt` (tabs, newline-terminated files); run `go fmt ./...` before every commit.
- Follow idiomatic Go naming: exported types/functions in CamelCase, unexported in mixedCaps; keep configuration keys in `snake_case` to match existing YAML.
- Keep functions small and side-effect aware; prefer dependency injection for external binaries (`xray`, `tor`).
- When adding linting, align with the README suggestion to adopt `golangci-lint` and include its config in `deployments` or `internal/`.

## Testing Guidelines
- Organize tests alongside code (for example `internal/users/store_test.go`).
- Use table-driven tests for config generation and user storage edge cases; mock filesystem interactions where practical.
- Ensure integration scripts validate generated Xray blobs before shipping; document any manual test steps in the PR.
- Aim for meaningful coverage on critical paths (`internal/xray`, `internal/api`) and keep `go test ./...` clean in CI.

## Commit & Pull Request Guidelines
- Write commits in imperative mood with concise subjects ("Add Tor health probe"), optionally prefixed using Conventional Commit tokens such as `feat:` or `fix:` when it clarifies scope.
- Squash noisy fixups locally; each commit should build and pass tests.
- In pull requests, link tracking issues when available, describe deployment impact, and attach command output or logs for risky changes.
- Capture manual verification steps (e.g., `make run` on sample config, docker-compose smoke test) so reviewers can reproduce quickly.

## Security & Configuration Tips
- Treat `deployments/docker/config.yaml` and `example/config.yaml` as templates; never commit real `xray.private_key` values.
- Use dedicated service accounts when applying `deployments/systemd/vpnd.service`, and document any capability changes (`setcap`) in your PR.
- Review container images and apt packages regularly; verify Tor reachability after every change to outbound networking.
