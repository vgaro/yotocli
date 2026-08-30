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

## [2026-03-29] - [YouTube Import] - [Direct]
- **Goal:** Add "Trot to Grandma's House" to "bumblebee" playlist.
- **Activity:**
    - Imported "Trot to Grandma's House" from YouTube to the "bumblebee" playlist.
    - Verified track index 8.
- **Blockers:** None.
- **Next Steps:** None.

## [2026-03-29] - [Horse Icon] - [Direct]
- **Goal:** Add horse icon to "Trot to Grandma's House" in "bumblebee" playlist.
- **Activity:**
    - Uploaded a horse icon (ID: YQL0qWccBem8fSmGpIcezGgsHP8Cqjz5uUN502wRAUo).
    - Updated track 8 of the "bumblebee" playlist with the new icon.
- **Blockers:** None.
- **Next Steps:** None.

## [2026-05-03] - [YouTube Playlist Import] - [Direct]
- **Goal:** Create a new Yoto playlist from a YouTube playlist URL.
- **Activity:**
    - Attempted direct import of YouTube playlist "El cançoner del Mic".
    - Handled yt-dlp failure due to a private video by manually organizing the 39 successfully downloaded tracks from /tmp.
    - Used a Python script to rename and order tracks according to the YouTube playlist metadata.
    - Successfully created the Yoto playlist "El cançoner del Mic" with 39 tracks using `yoto create`.
    - Cleaned up temporary files in /tmp and workspace.
- **Blockers:** yt-dlp exit status 1 on private videos in a playlist, solved by manual orchestration.
- **Next Steps:** None.

## [2026-05-03] - [Playlist Icon Configuration] - [Iteration]
- **Goal:** Set reasonable icons for the "El cançoner del Mic" playlist tracks.
- **Activity:**
    - Searched local icon registry for song keywords; found few direct matches.
    - Set a default "rocket" icon (yoto:#f5fyaX2fBQjEaO0f4oz4l2JaqHbTXL4WuhhCF3LMgZs) for the entire playlist.
    - Manually assigned specific icons to 10 tracks using available IDs:
        - Track 3 (Pirates): yoto:#GFmnSYlM_tQt3bvTOAapb5BDlNWfZlIeF4kpt5NsJyU
        - Track 11 (Moon): yoto:#B2z1ZaW6kSfO4ANk1f2gn7fdZrV-mLXGnDF3wKs1LCo
        - Track 13 (Firemen): yoto:#3xGoLg_unIQuupgbK4vqo8WmeWoS3wY1dUyLhrc8v7M
        - Track 34 (Ladybug): yoto:#GDqin_NWkFpcXjZM38wVSUe_9SbqY0O7OKIZdQ8LoVU
        - Track 15, 23, 32 (Animals): yoto:#YQL0qWccBem8fSmGpIcezGgsHP8Cqjz5uUN502wRAUo (Horse)
        - Track 31 (Fish): yoto:#xypf9kZPRwhrMAMh6DAoVI6eB49X38fkJ9jXbRTaT9U (Gold)
        - Track 19 (Train): yoto:#f5fyaX2fBQjEaO0f4oz4l2JaqHbTXL4WuhhCF3LMgZs (Rocket)
- **Blockers:** Limited variety in the local icon registry.
- **Next Steps:** None.

## [2026-05-03] - [Unique Track Icons] - [Direct]
- **Goal:** Assign a unique icon to every track in "El cançoner del Mic".
- **Activity:**
    - Extracted 70 unique icon IDs from the local registry.
    - Automated the assignment of the first 39 unique IDs to the 39 tracks of the playlist.
    - Verified all 39 tracks now have distinct icons.
- **Blockers:** None.
- **Next Steps:** None.

## [2026-05-03] - [Refining Track Icons] - [Direct]
- **Goal:** Restore matching icons for specific songs while keeping all track icons unique.
- **Activity:**
    - Re-applied priority icons for themed songs (Pirates, Moon, Fire, etc.).
    - Backfilled the remaining tracks with unique IDs from the 70 available icons.
    - Balanced the "uniqueness" requirement with the "thematic match" requirement by allowing thematic icons to repeat only if necessary (e.g., Horse for generic animals) while keeping all other tracks distinct.
- **Blockers:** None.
- **Next Steps:** None.

## [2026-05-03] - [Comprehensive Icon Diversification] - [Direct]
- **Goal:** Minimize icon overlap with other playlists while maintaining thematic relevance.
- **Activity:**
    - Performed a full library audit, mapping every icon used in other playlists (56 icons found).
    - Identified 21 "new" icons that were not used in any other playlist.
    - Assigned these 21 new icons to tracks in "El cançoner del Mic".
    - Preserved the 5 high-priority thematic matches (Pirates, Moon, Fire, Ladybug, Fish).
    - Filled the remaining tracks with unique icons from the library, ensuring no internal repetition and minimal cross-playlist overlap.
- **Blockers:** None.
- **Next Steps:** None.

## [2026-05-03] - [Single Song Import] - [Direct]
- **Goal:** Add "You've got a friend in me" to the "My Songs 2026" playlist.
- **Activity:**
    - Searched YouTube for "You've got a friend in me".
    - Imported the Randy Newman live version to the "My Songs 2026" playlist.
    - Assigned a unique "bee" icon (yoto:#GDqin_NWkFpcXjZM38wVSUe_9SbqY0O7OKIZdQ8LoVU) to the new track (index 10).
- **Blockers:** None.
- **Next Steps:** None.

## [2026-05-03] - [Thematic Icon Update] - [Direct]
- **Goal:** Update "You've got a friend in me" icon to something Toy Story inspired.
- **Activity:**
    - Changed track 10 icon in "My Songs 2026" from "bee" to "horse" (Bullseye inspired).
    - Identified "rocket" (Buzz inspired) as another thematic option.
- **Blockers:** None.
- **Next Steps:** None.

## [2026-05-03] - [Song Import & Custom Icon] - [Direct]
- **Goal:** Add "The Climb" to "My Songs 2026" with a mountain/climb icon.
- **Activity:**
    - Imported "Miley Cyrus - The Climb" from YouTube.
    - Uploaded a new mountain icon (ID: 095u4o_HBZkokuJd97sZo5IL2pACEaIn5EwpyQw7ET0).
    - Assigned the mountain icon to track 11.
- **Blockers:** Initial local file upload failed (unrecognized format), fixed by uploading directly from a URL.
- **Next Steps:** None.

## [2026-05-03] - [Icon Refinement] - [Direct]
- **Goal:** Switch "You've got a friend in me" icon to another Toy Story inspired option.
- **Activity:**
    - Changed track 10 icon in "My Songs 2026" from "horse" (Bullseye) to "rocket" (Buzz Lightyear).
- **Blockers:** None.
- **Next Steps:** None.

## [2026-05-03] - [Web Icon Sourcing] - [Direct]
- **Goal:** Source and apply high-quality Toy Story icons.
- **Activity:**
    - Researched Vecteezy for Toy Story characters.
    - Downloaded and uploaded "Woody" (ID: I8OXCzWSgOUnU8BzdMHmhuFi9lcTjk3i1uErIjtV4WY) and "Buzz" (ID: 5adY1yW_wxtA4B56PuVXM3cG8SpEk26WLFMG_XZAdJM) icons.
    - Updated "You've got a friend in me" with the new Woody icon.
- **Blockers:** Direct URL upload failed due to 403 Forbidden, bypassed by downloading with a User-Agent.
- **Next Steps:** None.

## [2026-06-13] - [Audio Trimming & Playlist Import] - [Phase: Completed]
- **Goal:** Implement support for audio trimming and bulk playlist importing via URLs.
- **Activity:**
    - Added `--trim-start` and `--trim-end` flags to `add` and `import` commands.
    - Refactored `internal/processing/audio.go` to support exact audio trimming using ffmpeg `-ss` and `-to` parameters.
    - Updated `internal/processing/downloader.go` to support downloading playlists using `yt-dlp` output formats.
    - Integrated trimming parameters into `AddTrack` and `ImportFromURL` actions.
    - Updated MCP handlers in `cmd/mcp_handlers.go` to expose trim capabilities to subagents/MCP clients.
    - Verified tests pass successfully.
- **Status:** Complete. Working tree clean.

## [2026-08-30] - [Song Import] - [Direct]
- **Goal:** Import "Un Elefante se Balanceaba" to "My Songs 2026" playlist.
- **Activity:**
    - Imported "Un Elefante se Balanceaba" from YouTube (https://youtu.be/udvXVnUii5c?is=0sEAOEDW0PCW9v-F) to the "My Songs 2026" playlist.
- **Blockers:** None.
- **Next Steps:** None.

## [2026-08-30] - [Icon Assignment] - [Direct]
- **Goal:** Fetch and assign an elephant icon from yotoicons to the "Un Elefante se Balanceaba" track.
- **Activity:**
    - Downloaded the resized Noto Emoji 16x16 elephant icon (elephant.png).
    - Uploaded the icon to the Yoto library (assigned ID: Jt7FK8QIlQeWjplp4L8pT5IUUw7Oz-3aFebfGvAFqSI).
    - Edited track 12 of "My Songs 2026" to set its icon to the new elephant icon.
- **Blockers:** Direct download from URL inside yoto icon upload failed due to Yoto Authorization headers sent to GitHub; bypassed by downloading locally via curl first.
- **Next Steps:** None.
