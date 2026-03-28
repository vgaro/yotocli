# yotocli Devlog

## [2026-03-05] - [Archive]
- **Goal:** Archive the `yotocli` project as it's being superseded or put into cold storage.
- **Activity:** Committed all pending changes (including `.gemini/` and `GEMINI.md`) and moved the project from `~/projects/yotocli` to `/home/mgaro/homelab/archive/yotocli`.
- **Blockers:** None.
- **Next Steps:** No further active development planned. This project is now in the lab's "abandoned" state for historical reference.

## [2026-03-15] - [Resurrection]
- **Goal:** Bring `yotocli` back to active development.
- **Activity:** Moved project from `archive/yotocli` to `workspace/yotocli`. Updated `lab-index.yaml` to `active_dev`. Successfully re-authenticated with Yoto.
- **Blockers:** None.
- **Next Steps:** Ready for testing or new features.

## [2026-03-16] - [Bulk Creation & Metadata]
- **Goal:** Process 98 stories from Google Drive into themed playlists with icons.
- **Activity:**
    - Downloaded 98 MP3s and 2 PNGs from Google Drive.
    - Organized files into 5 volumes (20 tracks each, except Vol. 5 with 18).
    - Fixed race condition in `internal/processing/audio.go` by using `os.CreateTemp`.
    - Enhanced `cmd/edit.go` with `--icon` flag.
    - Enhanced `cmd/ls.go` to display icon IDs in track details.
    - Batch-created 5 playlists: "El Tatano Vol. 1" to "El Tatano Vol. 5".
    - Gathered 12 unique pixel-art icon IDs from existing library and applied them to all 98 tracks via a batch script, ensuring variety.
- **Blockers:** Direct icon upload still failing with 400/403 (likely API policy for public clients), worked around by reusing existing library icon IDs.
- **Next Steps:** Explore player control features or library cleanup.

## [2026-03-28] - [YouTube Fixes & Titles] - [Maintenance]
- **Goal:** Resolve YouTube import failures and improve track naming.
- **Activity:**
    - Modified `internal/processing/downloader.go`: Added `--remote-components ejs:github` and `--js-runtimes node` to `yt-dlp` execution to fix "No supported JavaScript runtime" errors.
    - Updated `internal/actions/add.go`: Added an optional `title` parameter to `AddTrack` function.
    - Updated `internal/actions/import.go`: Modified `ImportFromURL` to pass the downloaded video title to `AddTrack`, ensuring imported tracks have descriptive names instead of temporary filenames.
    - Updated `cmd/add.go` and `cmd/mcp_handlers.go`: Adjusted call sites to match the new `AddTrack` signature.
    - Successfully imported "El ball dels pirates | El cançoner del Mic" to the **bumblebee** playlist.
    - Verified cleanup by removing the track with the temporary title.
- **Blockers:** None.
- **Next Steps:** Consider adding a flag to `add` command to allow specifying a custom track title during upload.
