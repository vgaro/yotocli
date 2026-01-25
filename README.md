# YotoCLI

A powerful, native command-line interface for managing your Yoto Player library. Built in Go for speed and portability.

## Features

- **🚀 One-Shot Creation:** Turn a folder of MP3s into a Yoto Playlist with a single command.
- **⚡ Parallel Uploads:** Uploads tracks concurrently for maximum speed.
- **🔊 Audio Normalization:** Automatically normalizes audio to -16 LUFS (Stereo) / -18 LUFS (Mono) using `ffmpeg`.
- **📂 File-System Like Management:** Manage your library like a filesystem (`ls`, `mv`, `cp`, `rm`).
- **🛠️ Advanced Editing:** Reorder tracks, move tracks between playlists, and append new files easily.

## Installation

### Prerequisites
- **Go 1.24+** (to build)
- **ffmpeg** (for audio normalization)

### Build
```bash
git clone https://github.com/vgaro/yotocli.git
cd yotocli
go build -o yoto main.go
sudo mv yoto /usr/local/bin/
```

## Usage

For detailed information on every command, see the [Command Documentation](docs/commands/yoto.md).

### 1. Authentication
First, log in to your Yoto account. This uses the secure Device Code flow.
```bash
yoto login
```

### 2. Creating a Card (The "Happy Path")
Create a new playlist from a local directory of audio files.
```bash
# Creates a playlist named "Bedtime Stories" with all audio files in the folder
yoto create --name "Bedtime Stories" ./path/to/mp3s/

# Disable normalization if files are already processed
yoto create --no-normalize ./path/to/mp3s/
```

### 3. Listing Content
List all playlists or deep-dive into tracks.
```bash
# List all playlists
yoto ls

# List tracks in a specific playlist (fuzzy name matching supported)
yoto ls "Bedtime"

# Show details of a specific track (by index)
yoto ls "Bedtime/1"
```

### 4. Downloading Content
Backup your library or extract tracks.
```bash
# Download entire playlist (creates folder "Bedtime Stories")
yoto download "Bedtime Stories"

# Download to a specific backup folder
yoto download "Bedtime Stories" ./backups/

# Download a single track
yoto download "Bedtime Stories/1"
```

### 5. Editing Content
Modify titles and metadata.
```bash
# Rename a playlist
yoto edit "Bedtime Stories" --name "Sleepy Time"

# Change author and description
yoto edit "Sleepy Time" --author "Dad" --description "Read by Dad"

# Rename a specific track
yoto edit "Sleepy Time/1" --name "Chapter 1"
```

### 6. Importing from Web (YouTube/etc)
Download audio directly from the web to a card.
```bash
# Import video to a specific playlist
yoto import "https://youtu.be/..." --playlist "Bedtime Stories"

# Import to a new playlist (uses video title)
yoto import "https://youtu.be/..."
```

### 7. Device Control
Check your player's status.
```bash
# Check battery and current activity
yoto status
```

### 8. Managing Playlists
Treat your library like a filesystem.

**Add a track:**
```bash
# Append to end
yoto add "Bedtime" ./new-story.mp3

# Insert at specific position (e.g., position 2)
yoto add "Bedtime/2" ./intro.mp3
```

**Remove content:**
```bash
# Remove a specific track
yoto rm "Bedtime/2"

# Remove entire playlist
yoto rm "Bedtime"
```

**Reorder / Move:**
```bash
# Move track up or down
yoto mvup "Bedtime/2"
yoto mvdown "Bedtime/1"

# Move track to position 5
yoto mv "Bedtime/1" "Bedtime/5"

# Move track to ANOTHER playlist
yoto mv "Bedtime/1" "Dance Party/"
```

**Copy:**
```bash
# Duplicate a track to another playlist
yoto cp "Bedtime/1" "Favorites/"
```

## Configuration
Configuration is stored in `~/.config/yotocli/config.yaml`.

### Custom Client ID
If you want to use your own Yoto API Client ID, add it to the config.
You can get one from the **Yoto Developer Dashboard**: https://dashboard.yoto.dev/

```yaml
auth:
  client_id: "YOUR_CLIENT_ID"
```

## Troubleshooting
- **Normalization Failed:** Ensure `ffmpeg` is installed and in your PATH.
- **Authentication:** If commands fail with 401/403, run `yoto login` again.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
