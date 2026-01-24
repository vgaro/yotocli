# YotoCLI Architecture

This document describes the high-level architecture of `yotocli`.

## 1. Design Philosophy: Filesystem Abstraction

The core concept of `yotocli` is treating the Yoto Cloud Library as a local filesystem.
- **Root:** The user's card library.
- **Directory:** A specific Card/Playlist.
- **File:** A Track within a playlist.

This allows us to map familiar commands (`ls`, `mv`, `cp`, `rm`) directly to API actions.
- `ls "Bedtime Stories"` -> Lists contents of "directory".
- `mv "Bedtime Stories/Track 1" "Favorites/"` -> Moves file between "directories".

## 2. Project Structure

The project follows a standard Go CLI layout:

- **`cmd/`**: The entry point for all commands (using `Cobra`).
    - Handles flag parsing, user input validation, and output formatting.
    - *No heavy logic here.* It delegates to `pkg/yoto` or `internal/utils`.

- **`pkg/yoto/`**: The Core API Client.
    - Wraps the Yoto HTTP API (unofficial/reverse-engineered).
    - **Models:** Defines `Card`, `Chapter`, `Track` structs mirroring the JSON response.
    - **Auth:** Handles OAuth2 Device Flow and Token Refresh.
    - **Upload:** Manages the multi-step upload (Get URL -> PUT -> Poll Transcode).
    - *Zero dependency on CLI logic.* Can be imported by other Go programs.

- **`internal/utils/`**: Shared helpers.
    - **`fs.go`**: Filesystem safety (Sanitization).
    - **`finder.go`**: Logic for the "Slash Syntax" (`Playlist/Track` parsing).
    - **`playlist_utils.go`**: Logic for reordering/renumbering playlist arrays.

- **`internal/processing/`**: Audio processing.
    - Wraps `ffmpeg` calls for normalization.
    - Wraps `yt-dlp` for downloading audio from external URLs.

- **`internal/config/`**: Configuration management.
    - Uses `Viper` to load/save tokens in `~/.config/yotocli/config.yaml`.

## 3. Key Workflows

### Import (Web to Yoto)
1.  **Download:** `cmd/import` calls `yt-dlp` to fetch audio (best quality) -> converts to MP3.
2.  **Normalize:** Runs `ffmpeg` on the downloaded file.
3.  **Upload/Add:** Reuses the standard Upload -> Transcode -> Add Track flow.

### Upload & Creation
1.  **Scan:** `cmd/create` scans a local directory.
2.  **Normalize:** `internal/processing` runs `ffmpeg` to target -16 LUFS.
3.  **Upload:** Files are uploaded in parallel (concurrency limit: 5) to Yoto's S3 bucket.
4.  **Transcode:** The CLI polls the API until Yoto finishes processing.
5.  **Create:** A `POST /content` request creates the card with the new track references.

### Authentication
Uses the **OAuth2 Device Authorization Flow**.
1.  CLI requests a code (`POST /oauth/device/code`).
2.  User visits URL and enters code.
3.  CLI polls (`POST /oauth/token`) until authorized.
4.  Tokens are saved securely to config.

## 4. API Notes
The API endpoints used are based on reverse-engineering the Yoto Web/App traffic.
- **Base URL:** `https://api.yotoplay.com`
- **Auth URL:** `https://login.yotoplay.com`
- **Content:** `POST /content` (Create), `PATCH /content/{id}` (Update), `DELETE /content/{id}` (Delete).
