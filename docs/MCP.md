# Yoto MCP Server Guide

This guide explains how to use the Yoto CLI as a Model Context Protocol (MCP) server. This allows AI assistants (like Claude Desktop) to directly interact with your Yoto library.

## Prerequisites

1.  **YotoCLI Installed:** Ensure you have the `yoto` binary installed and in your PATH.
2.  **Authenticated:** Run `yoto login` at least once to authenticate.

## Configuration for Claude Desktop

To use Yoto with Claude Desktop, add the following to your `claude_desktop_config.json` file.

**Location:**
- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

**Configuration:**

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
*Replace `/absolute/path/to/yoto` with the actual path (e.g., `/usr/local/bin/yoto` or `/home/user/go/bin/yoto`).*

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
- **Input:** `url` (string), `playlist_name` (optional - creates new if empty), `no_normalize` (boolean, optional)

### `add_track`
Uploads a local audio file to a playlist.
- **Input:** `file_path` (string), `playlist_name` (string), `no_normalize` (boolean, optional)

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
