# Contributing to wcheck

Thank you for your interest in improving `wcheck`! We welcome contributions from the community.

## Project Structure

- `cmd/`: CLI command definitions (Cobra).
- `internal/engine/`: Core headless browser logic and interaction monkey.
- `internal/crawler/`: Link discovery and site scouting.
- `internal/reporter/`: Pretty CLI reporting and table generation.
- `tests/`: Integration tests for scanner logic.

## Workflow

1.  **Fork and Clone:** Create a feature branch.
2.  **Go Version:** Ensure you are using Go 1.23+.
3.  **Run Tests:** Always verify your changes before submitting:
    ```bash
    go test ./... -v
    go test ./tests/... -v
    ```
4.  **Linting:** Follow standard Go idioms (`gofmt`).
5.  **Submit PR:** Provide a clear description of the fix or feature.

## Development Tips

- Use the `--verbose` flag when debugging interaction monkey logic.
- Use `httptest` for mocking web servers in new tests.
- Always check for potential race conditions when modifying parallel worker logic.

## Feedback

Found a bug or have a feature request? Please open an issue on GitHub.
