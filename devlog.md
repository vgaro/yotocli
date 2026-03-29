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

## [2026-03-28] - [Icon Search & Custom Titles] - [Iteration]
- **Goal:** Enable local icon searching and custom titles for local uploads.
- **Activity:**
    - **Icon Registry**: Implemented `internal/config/icons.go` to manage a local `icons.yaml` mapping ID to Name/Tags.
    - **Icon Search**: Added `yoto icon ls` and `yoto icon search <query>` to list and search the local registry.
    - **Icon Metadata**: Updated `yoto icon upload` to accept `--name` and `--tags`, automatically saving them to the registry.
    - **Custom Titles**: Added `--title` (-t) flag to `yoto add` command to allow custom track names during local file upload.
    - **MCP Update**: Exposed `list_icons` and `search_icons` via the MCP server and updated `upload_icon` to support names/tags.
    - **Refactor**: Cleaned up `pkg/yoto/client.go` structural issues and improved error handling for icon listing.
    - **Testing**: Added unit tests for icon registry logic and search functionality in `internal/config/icons_test.go` and `internal/actions/icon_test.go`.
- **Blockers:** Yoto API does not provide icon names, necessitating the local registry workaround.
- **Next Steps:** Maintain the local icon registry as new icons are uploaded.

## [2026-03-29] - [YouTube & Rocket Icon] - [Direct]
- **Goal:** Add YouTube video to "bumblebee" playlist with a rocket icon.
- **Activity:**
    - Imported "David Bowie - Space Oddity (Official Video)" from YouTube to the "bumblebee" playlist.
    - Uploaded a new rocket icon from img.icons8.com (ID: f5fyaX2fBQjEaO0f4oz4l2JaqHbTXL4WuhhCF3LMgZs) as the official library ID 'yoto:rocket' was rejected by the API.
    - Updated track 7 of the "bumblebee" playlist with the new rocket icon.
- **Blockers:** Yoto API rejected 'yoto:rocket' ID (bad-request), solved by uploading a custom icon and using its generated ID.
- **Next Steps:** None.
