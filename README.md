[![Go Report Card](https://goreportcard.com/badge/github.com/mschneider82/sharecmd)](https://goreportcard.com/report/github.com/mschneider82/sharecmd) [![GoDoc](https://godoc.org/github.com/mschneider82/sharecmd?status.svg)](https://godoc.org/github.com/mschneider82/sharecmd)

![gopher](gopher.png)

# Go Share files!

Share your files with your friends using cloud providers with just one command.

# Supported Cloud Providers:

* **HTTP Upload** — any server that accepts HTTP PUT (e.g. WebDAV, custom APIs)
* Box
* Dropbox
* Google Drive
* OpenDrive
* Seafile (also private hosted)
* Nextcloud / Owncloud
* Any missing? Create an Issue or PR!

# How to share?

```
$ share somedocument.pdf
Uploading somedocument.pdf

████████████████████████████████████████  100%
361 B / 361 B  120.3 KiB/s

URL: https://drive.google.com/file/d/1C77TZ.../view?usp=sharing
URL copied to clipboard!
```

# How to setup?

```
$ share --setup
```

An interactive TUI guides you through the setup. You can configure multiple providers and switch between them.

```
ShareCmd Setup  (active: my-dropbox)
? What would you like to do?
  > Select active provider
    Add new provider
    Edit provider
    Delete provider
    Preferences
    Quit
```

## Multiple providers

You can add as many provider configurations as you want, each with a unique label (e.g. `work-nextcloud`, `personal-dropbox`). Use **Select active provider** to switch between them.

## Preferences

Under **Preferences** you can toggle:

* **Copy URL to clipboard** — enabled by default
* **QR code display** — enabled by default

# How to install?

[Download precompiled binaries](https://github.com/mschneider82/sharecmd/releases) for your OS
or compile it from source.

**Using Go:**
```
go install schneider.vip/share@latest
```

**Using Snap (Linux):**
```
snap install share
```

**Using Homebrew (macOS/Linux):**

```
brew tap mschneider82/sharecmd https://github.com/mschneider82/sharecmd
brew install sharecmd
```

# CLI Usage

```
$ share [flags] [file] [provider]
```

| Flag | Description |
|------|-------------|
| `--setup`, `-s` | Launch interactive setup |
| `--select`, `-p` | Select provider for this upload interactively |
| `--version`, `-v` | Print version and exit |
| `--config PATH` | Path to config file (default: `~/.config/sharecmd/config.json`) |

If no active provider is configured, setup launches automatically.

## Provider Override

You can temporarily override the active provider by specifying its label as an argument. The order of arguments doesn't matter:

```
$ share filename.zip dropbox      # Upload to dropbox
$ share dropbox filename.zip      # Same as above
$ share my-nextcloud report.pdf   # Upload to provider labeled "my-nextcloud"
```

If an argument matches a configured provider label, that provider is used instead of the default active provider.

**Interactive provider selection:**

Use the `--select` (or `-p`) flag to choose a provider interactively:

```
$ share --select document.pdf
? Select provider for this upload
  > my-dropbox (dropbox)
    work-nextcloud (nextcloud)
    personal-gdrive (googledrive)
```

If you provide an invalid provider name, an interactive menu appears automatically:

```
$ share dr document.pdf
Argument "dr" is not a configured provider.
? Select provider for this upload
  > my-dropbox (dropbox)
    work-nextcloud (nextcloud)
    personal-gdrive (googledrive)
```

# Notes

ShareCmd uploads the file to the configured cloud provider and creates a public
share link for anyone who has the link. The link is printed (with optional QR code)
and optionally copied to the system clipboard (Windows/Linux/macOS).

The configuration is stored in `~/.config/sharecmd/config.json`. Old single-provider
configs (v1) are automatically migrated to the new multi-provider format on first load.

# Provider Notes

## HTTP Upload
Uploads files via HTTP PUT to a configurable base URL. The file is PUT to `<base-url>/<filename>`.

Custom HTTP headers can be added as JSON. Header values support Go template functions for dynamic values:

| Function | Description | Example output |
|----------|-------------|----------------|
| `{{now "2006-01-02"}}` | Current date | `2026-02-19` |
| `{{addDays N "2006-01-02"}}` | Date + N days | `2026-02-26` |

Example header configuration for a purge-date header one week in the future:
```json
{"x-purgets": "{{addDays 7 \"2006-01-02\"}}T16:09:09+02:00"}
```

## Box
Uploads all files to `/sharecmd` (folder auto-generated). Re-uploading the same filename creates a new version.

## Dropbox
Uploads all files to `/` (overwrite mode).

## Google Drive
Uploads all files to `/sharecmd` (folder auto-generated).

## OpenDrive
Uploads all files to `/sharecmd` (folder auto-generated).

## Seafile
Creates a new library called `sharecmd` on setup.

## Nextcloud / Owncloud
The folder `/sharecmd` is auto-generated.
