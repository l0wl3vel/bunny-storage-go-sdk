# 🐇 Bunny.net Edge Storage Go Library

This is a new Library for using the Bunny.net Object Storage Service in Go. It has been heavily inspired by James Pond`s [bunnystorage-go](https://sr.ht/~jamesponddotco/bunnystorage-go/).


## ❓ Why another library?

I wrote this because of a open PR implementing Bunny.net as Object Storage in JuiceFS and the maintainers were a bit hesitant merging code with many homebrew dependencies. This is using just three widespread dependencies (resty, uuid, logrus) and is about 100 lines of library code, excluding tests.

The E2E test coverage is currently at about 70%

This library features a simple API and allows you to focus on what`s important


## 🦾 Getting Started

```go
import "net/url"
import "github.com/l0wl3vel/bunnystorage-go"

endpoint, err := endpoint.Parse("https://la.storage.bunnycdn.com/mystoragezone/")
if err != nil	{
    panic(err)
}
bunnyclient = bunnystorage.NewClient(endpoint, password)

content := make([]byte, 1048576)
// Fill content with data

// The last argument controls if a checksum is included in the request
err := bunnyclient.Upload("foo/bar.txt", content, true) 
if err != nil 	{
	panic(err)
}

```

## 🤔 Further Ideas

[ ] Implement Pull Zone support


# ❤️ Thanks to

- James Pond for creating the original [bunnystorage-go](https://sr.ht/~jamesponddotco/bunnystorage-go/) library, which allowed me to start prototyping my work on JuiceFS immediately 
- Bunny.net for creating an awesome performing and attractively priced object storage solution

# Disclaimer

This is a community implementation of the Bunny.net Storage API. It is not sponsored or endorsed by BunnyWay d.o.o.