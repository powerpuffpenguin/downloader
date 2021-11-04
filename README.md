# downloader
file downloader library

downloader is a download program and can be used as a download library

currently only supports http download

features:
* Support breakpoint resume download
* Support validation hash

# as program

```
$ ./downloader http -h
http download

Usage:
  downloader http [flags]

Examples:
  downloader http http https://ww.google.com
  downloader http https://ww.google.com/1 http https://ww.google.com/2
  downloader http -n file1 https://ww.google.com/1 http https://ww.google.com/2

Flags:
  -c, --check string     checksum function ['MD4','MD5','SHA1','SHA224','SHA256','SHA384','SHA512','MD5SHA1','RIPEMD160','SHA3_224','SHA3_256','SHA3_384','SHA3_512','SHA512_224','SHA512_256','BLAKE2s_256','BLAKE2b_256','BLAKE2b_384','BLAKE2b_512']
  -H, --header strings   request header (default [User-Agent=Downloader/v1.0.0 (linux amd64 go1.16.5)])
  -h, --help             help for http
  -j, --json             use json encoding to download the status file
  -n, --names strings    download saved filename
  -s, --sum strings      hash sum hex string
      --sync int         whenever the specified length of data is downloaded, the download status is synchronized (default 5242880)
```

# as library
```
package main

import (
	"log"

	"github.com/powerpuffpenguin/downloader/http"
)

func main() {
	// Set optional parameters
	opts := []http.Option{
		http.WithJSON(true),
		http.WithSync(1024 * 1024 * 5),
	}
	// Create a download task
	w := http.New(
		`url`, // Download URL
		`dst`, // File name saved locally
		opts...,
	)
	// Execute download
	e := w.Serve()
	if e != nil {
		log.Fatalln(e)
	}
}
```