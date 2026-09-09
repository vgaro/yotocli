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
    - Wraps `yt-dlp` for downloading audio from external URLs.

- **`internal/config/`**: Configuration management.
    - Uses `Viper` to load/save tokens in `~/.config/yotocli/config.yaml`.

## 3. Key Workflows

### Import (Web to Yoto)
1.  **Download:** `cmd/import` calls `yt-dlp` to fetch audio (best quality) -> converts to MP3. A URL holding several items (a playlist, an RSS feed) yields one file per item.
2.  **Upload/Add:** Hands the files to `actions.AddTracks`, which is the same path `add` and `create` use.

### Upload & Creation
Every command that puts audio on a card goes through `actions.AddTracks`:
1.  **Resolve the playlist:** Once per batch, `GET /card/family/library` then `GET /content/{id}`, or a new card if the name is not there. Doing this once is what stops a batch of uploads from each creating its own copy of the same playlist.
2.  **Upload:** Files are uploaded concurrently (concurrency limit: 10) to Yoto's S3 bucket. This is the only concurrent step, and it never touches the card.
3.  **Transcode:** The CLI polls the API until Yoto finishes processing. Yoto normalizes the audio to -16 LUFS and re-encodes it to Opus here, which is why the CLI does no audio processing of its own.
4.  **Write:** One `PUT`/`POST /content` adds every new chapter at once, so concurrent uploads cannot drop each other's tracks. Nothing is written unless every upload succeeded.

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
