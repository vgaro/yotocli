# Contributing to YotoCLI

Thank you for helping improve YotoCLI!

## Development Setup

1. **Go:** Ensure you have Go 1.22+ installed.
2. **FFmpeg:** Install `ffmpeg` and `ffprobe` for audio features.
3. **Dependencies:** Run `go mod download`.

## Running Tests

We use standard Go testing.

### Unit Tests
Tests that don't require the network or external binaries:
```bash
go test ./internal/utils/...
```

### Integration Tests
Tests that use a mock Yoto API server:
```bash
go test ./pkg/yoto/...
```

### Full Suite
```bash
go test ./...
```

## Adding Commands

1. Create a new file in `cmd/`.
2. Use the `rootCmd` to register your command.
3. Use `apiClient` (globally available in the `cmd` package) to interact with Yoto.
4. Add an `Example` section to your Cobra command definition.

## Coding Standards
- **Layering:** Keep CLI logic in `cmd/` and API logic in `pkg/yoto/`.
- **Error Handling:** Return errors from `pkg/` and handle them (print/exit) in `cmd/`.
- **Sanitization:** Always use `utils.SanitizeFilename` when creating local files from API data.
