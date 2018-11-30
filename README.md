![gopher](gopher.png)

# Go Share files!

Share your files with your friends using Cloudproviders with just one command.

# Supported Cloud Providers:

* Dropbox
* Google Drive
* Seafile (also private hosted)
* Nextcloud / Owncloud
* Any missing? Create an Issue or PR!

# Howto share?

```
user@srv# share /etc/hosts
Uploading 361 B/361 B
URL: https://drive.google.com/open?id=1C77TZBMT0PESUvsIPetGzrK36LqGFqza
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

On MacOSX you can use brew

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

## Seafile:
It creates a new Library called sharecmd on setup

# Documentation:
[![GoDoc](https://godoc.org/github.com/mschneider82/sharecmd?status.svg)](https://godoc.org/github.com/mschneider82/sharecmd)
