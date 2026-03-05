# YotoCLI

## Project Overview

**YotoCLI** is a native Go command-line interface for managing Yoto Player audio libraries. It treats the Yoto Cloud Library as a filesystem, enabling users to manage playlists and tracks using familiar commands like `ls`, `cp`, `mv`, and `rm`.

It provides advanced features such as:
*   **One-Shot Creation:** Converting local directories of audio files into playlists.
*   **Parallel Uploads:** Concurrent uploading for speed.
*   **Audio Normalization:** Automatic audio processing using `ffmpeg`.
*   **Web Import:** Downloading and importing audio directly from URLs (e.g., YouTube).
*   **Filesystem Metaphor:** Intuitive command structure mimicking standard shell file operations.

## Architecture

The project adheres to the Standard Go Project Layout:

*   **`cmd/`**: Contains the main entry points for the CLI commands. Uses [Cobra](https://github.com/spf13/cobra) for command handling.
    *   This layer handles user input, flag parsing, and output formatting. It delegates business logic to `pkg` and `internal`.
*   **`pkg/yoto/`**: The core API client library.
    *   Encapsulates all Yoto API interactions (Device Auth, Content Management, Uploads).
    *   Designed to be reusable in other Go applications.
*   **`internal/`**: Private application code.
    *   `config/`: Configuration management using [Viper](https://github.com/spf13/viper).
    *   `processing/`: Audio processing logic (wrappers for `ffmpeg` and `yt-dlp`).
    *   `utils/`: Shared utilities for filesystem operations, path parsing (Slash Syntax), and playlist manipulation.
*   **`docs/`**: Documentation files, including architecture notes and individual command references.

## Building and Running

### Prerequisites
*   **Go 1.24+**
*   **ffmpeg** (Required for audio normalization)
*   **yt-dlp** (Optional, for `import` command)

### Build Commands (Makefile)

The project includes a `Makefile` to simplify common tasks:

*   **Build Binary:**
    ```bash
    make build
    # Output: ./yoto
    ```
*   **Run Tests:**
    ```bash
    make test
    # Runs all tests in the project
    ```
*   **Run Tests (Verbose):**
    ```bash
    make test-v
    ```
*   **Generate Documentation:**
    ```bash
    make docs
    # Generates Markdown docs in docs/commands/
    ```
*   **Install:**
    ```bash
    make install
    # Builds and moves binary to /usr/local/bin/
    ```

## Development Conventions

*   **Layering:** Maintain strict separation between CLI logic (`cmd/`) and business logic (`pkg/`, `internal/`).
    *   `cmd/` should only handle "how" the command is invoked.
    *   `pkg/yoto/` should handle "what" the API does.
*   **Error Handling:** Errors should be returned by functions in `pkg/` and `internal/`, and handled (logged/displayed) in `cmd/`.
*   **Testing:**
    *   **Unit Tests:** Place in `internal/` for logic that doesn't require external services.
    *   **Integration Tests:** Place in `pkg/yoto/` to test API interactions (using mocks where appropriate).
*   **Sanitization:** Always use `utils.SanitizeFilename` when dealing with file paths derived from user input or API responses.

## Key Files

*   `main.go`: The application entry point.
*   `cmd/root.go`: Configuration of the root command and global flags.
*   `internal/config/manager.go`: Handles loading and saving `config.yaml`.
*   `pkg/yoto/client.go`: Main struct for the Yoto API client.
*   `docs/ARCHITECTURE.md`: Detailed architectural design notes.
