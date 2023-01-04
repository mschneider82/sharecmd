[![Go Report Card](https://goreportcard.com/badge/github.com/mschneider82/sharecmd)](https://goreportcard.com/report/github.com/mschneider82/sharecmd) [![GoDoc](https://godoc.org/github.com/mschneider82/sharecmd?status.svg)](https://godoc.org/github.com/mschneider82/sharecmd)

![gopher](gopher.png)

# Go Share files!

Share your files with your friends using Cloudproviders with just one command.

# Supported Cloud Providers:

* Dropbox
* Google Drive (currently not working)
* OpenDrive
* Seafile (also private hosted)
* Nextcloud / Owncloud
* Any missing? Create an Issue or PR!

# Optional Support for URL Shortener:

* Biturl.top
* ...

# Howto share?

```
user@srv# share somedocument.pdf
Uploading 361 B/361 B
URL: https://drive.google.com/open?id=1C77TZBMT0PESUvsIPetGzrK36LqGFqza
Short URL: https://biturl.top/67vE32
URL copied to clipboard!
```

# Howto setup?

```
user@srv# share --setup
```
Select a provider and connect the app to your account. The token will be saved to your disk.

# Howto install?

[Download precompiled binarys](https://github.com/mschneider82/sharecmd/releases) for your OS
or compile it from source.

On Linux/Mac/Windows you can use this curl command to always get the latest version:

```
curl -sfL https://install.goreleaser.com/github.com/mschneider82/sharecmd.sh | sh
```

Or if you use MacOSX you can also use brew:

```
brew install https://github.com/mschneider82/sharecmd/raw/master/Formula/sharecmd.rb
```

# Notes:
Sharecmd uploads the file to the configured cloud provider and does a public
share of the file for anyone who has the link. The link will be copyed to system
clipboard (windows/linux/macos)

# Provider Notes:

## Dropbox:
It uploads all files to /Apps/sharecmd (folder auto generated)

## Googledrive:
It uploads all files to /sharecmd (folder auto generated)

## Opendrive
It uploads all files to /sharecmd (folder auto generated)

## Seafile:
It creates a new Library called sharecmd on setup

## Own/Nextcloud:
The folder /sharecmd will be auto generated.
