![gopher](gopher.png)

# Go Share files!

Share your files easily with friends using Cloudproviders.

# Supported Cloud Providers:

* Dropbox
* Google Drive

# Howto share?

```
user@srv# sharecmd /etc/hosts
Uploading 361 B/361 B
2018/11/28 19:03:07 URL: https://drive.google.com/open?id=1C77TZBMT0PESUvsIPetGzrK36LqGFqza
URL copied to clipboard!
```

# Howto setup?

```
user@srv# sharecmd --setup
```

# Howto install?

[Download precompiled binarys](https://github.com/mschneider82/sharecmd/releases) for your OS
or compile it from source.

# Notes:
Sharecmd uploads the file to the configured cloud provider and does a public
share of the file for anyone who has the link. The link will be copyed to system
clipboard (windows/linux/macos)

# Provider Notes:

## Dropbox:
It uploads alls files to /Apps/sharecmd (folder auto generated)

## Googledrive:
It uploads alls files to /sharecmd (folder auto generated)


# Documentation:
[![GoDoc](https://godoc.org/github.com/mschneider82/easygo?status.svg)](https://godoc.org/github.com/mschneider82/easygo)