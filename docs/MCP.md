# Yoto MCP Server Guide

This guide explains how to use the Yoto CLI as a Model Context Protocol (MCP) server. This allows AI assistants (like Claude Desktop) to directly interact with your Yoto library.

## Prerequisites

1.  **YotoCLI Installed:** Ensure you have the `yoto` binary installed and in your PATH.
2.  **Authenticated:** Run `yoto login` at least once to authenticate.

## Configuration for Claude Desktop

### Option 1: Local (Stdio)
Use this if `yoto` is installed on the same computer as Claude.

```json
{
  "mcpServers": {
    "yoto": {
      "command": "/absolute/path/to/yoto",
      "args": ["mcp"]
    }
  }
}
```

### Option 2: Remote (SSE / Tailscale)
Use this if `yoto` is running on a server (e.g., Homelab) and you want to access it from your laptop.

**Server (Docker - Recommended):**
1.  Ensure `docker-compose.yml` is present.
2.  Run: `docker-compose up -d`
3.  Server will be available at `http://<YOUR_SERVER_IP>:8082/sse`

**Server (Manual):**
```bash
yoto mcp --transport sse --port 8080
```

**Claude Config:**
```json
{
  "mcpServers": {
    "yoto-remote": {
      "url": "http://<YOUR_SERVER_IP>:8082/sse"
    }
  }
}
```

## Available Tools

The MCP server exposes the following tools to the AI:

### `list_playlists`
Lists all playlists (cards) in your library.
- **Returns:** List of `{id, title}`.

### `get_playlist`
Gets detailed information about a specific playlist, including tracks.
- **Input:** `playlist_id` (string)
- **Returns:** Full Card object.

### `list_devices`
Lists all registered Yoto players.
- **Returns:** List of devices.

### `get_device_status`
Checks the status of a specific player (battery, volume, online status).
- **Input:** `device_id` (string)
- **Returns:** Status object.

### `create_playlist`
Creates a new empty playlist.
- **Input:** `title` (string), `description` (optional), `author` (optional)

### `delete_playlist`
Deletes a playlist.
- **Input:** `playlist_id` (string)

### `edit_playlist`
Updates the metadata of a playlist.
- **Input:** `playlist_id` (string), `title` (optional), `description` (optional), `author` (optional)

### `import_from_url`
Downloads audio from a URL (e.g., YouTube), normalizes it, and adds it to a playlist.
- **Input:** `url` (string), `playlist_name` (optional - creates new if empty or not found), `no_normalize` (boolean, optional)

### `add_track`
Uploads a local audio file to a playlist.
- **Input:** `file_path` (string), `playlist_name` (string - creates new if not found), `icon_id` (string, optional), `no_normalize` (boolean, optional)

### `set_track_icon`
Sets the icon for a specific track in a playlist.
- **Input:** `playlist_id` (string), `track_index` (integer, 1-based), `icon_id` (string - e.g., "yoto:#HASH" or just the hash)

### `upload_icon`
Uploads a custom icon (PNG/GIF) to your library. Returns the Icon ID.
- **Input:** `file_path` (string - local path or URL)

### `remove_track`
Removes a specific track from a playlist.
- **Input:** `playlist_id` (string), `track_index` (integer, 1-based)

### `move_track`
Reorders a track within a playlist or moves it to another.
- **Input:** `playlist_id` (string), `track_index` (integer, 1-based), `new_position` (integer, 1-based), `dest_playlist_id` (optional)

### `copy_track`
Copies a track to another playlist.
- **Input:** `playlist_id` (string), `track_index` (integer, 1-based), `dest_playlist_id` (optional), `new_position` (integer, 1-based)

### `set_volume`
Sets the volume of a Yoto player.
- **Input:** `volume` (integer, 0-100), `device_id` (optional)

### `play_card`
Starts playing a playlist on a device.
- **Input:** `playlist_id` (string), `device_id` (optional)

### `stop_player`
Stops playback on a device.
- **Input:** `device_id` (optional)

### `pause_player`
Pauses playback on a device.
- **Input:** `device_id` (optional)

## Common Workflows

### Creating a Custom Card
1.  **Create:** `create_playlist(title="My Story", author="Me")`
2.  **Add Content:**
    *   From YouTube: `import_from_url(url="...", playlist_name="My Story")`
    *   From Local File: `add_track(file_path="/tmp/story.mp3", playlist_name="My Story")`
3.  **Add Icon:**
    *   Upload: `id = upload_icon(file_path="https://.../icon.png")`
    *   Set: `set_track_icon(playlist_id=..., track_index=1, icon_id=id)`

### Managing Playback
1.  **Check Status:** `get_device_status()` to see what's playing.
2.  **Control:** `play_card()`, `pause_player()`, `set_volume()`.

## Troubleshooting

-   **"Unauthorized" Error:** The access token has expired and the server hasn't refreshed it yet, or the configuration is stale.
    *   **Fix:** Run `yoto ls` in your terminal to force a refresh, then **restart the MCP server** (e.g., restart Claude Desktop).
-   **"Track not found" / "Invalid index":** Remember that `track_index` is 1-based (matches the Yoto app UI), not 0-based.
-   **"yt-dlp not found":** The `import_from_url` tool requires `yt-dlp` to be installed on the host system.

## Example Prompts

Once configured, you can ask Claude things like:

- "List my Yoto playlists."
- "What is the battery level of the Kids' Room Yoto?"
- "Set the volume on the Yoto to 15."
- "Play the 'Bedtime Stories' playlist."
- "Pause the music."
- "Create a new playlist called 'Sleepy Time' with author 'Dad'."
- "Rename the 'Sleepy Time' playlist to 'Bedtime'."
- "Add this video https://youtube.com/... to my 'Bedtime' playlist."
- "Show me the tracks in the 'Bedtime Stories' playlist."
