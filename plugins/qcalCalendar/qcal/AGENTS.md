# qcal Agent Guidelines

## Build/Lint/Test Commands
- Build: `make` (compiles to `qcal` binary)
- Clean: `make clean`
- Install: `sudo make install`
- Cross-platform: `make darwin` (macOS), `make linux-arm`, `make windows`
- Lint: `gofmt -l .` (check formatting), `go vet ./...` (static analysis)
- Test: `go test ./...` (run all tests), `go test -run TestName` (run single test)

## Code Style Guidelines
- Imports: Standard library first (alphabetical), third-party second, blank line between groups.
- Formatting: Use `gofmt`, 4-space indentation, reasonable line length.
- Types & Naming: Structs PascalCase (e.g., `Event`), functions camelCase (e.g., `fetchCalData`), constants ALL_CAPS (e.g., `IcsFormat`), variables camelCase.
- Error Handling: `log.Fatal(err)` for fatal, `checkError(e)` or return for non-fatal, check `resp.Status` for HTTP.
- General: Minimal comments, receiver methods for structs (e.g., `(e Event)fancyOutput()`), global configs OK for CLI, use `time.Time` for dates.