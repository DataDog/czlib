# Disclaimer
*This repository is considered stable but not actively maintained anymore. It is still in use in many places and safe for production use; but the zlib protocol being stable, we have not made any changes in recent times. Time to reply on issues/PRs may not be on par with other Datadog's repositories.*


# czlib

[![GoDoc](https://godoc.org/github.com/DataDog/czlib?status.svg)](https://godoc.org/github.com/DataDog/czlib)

`czlib` started as a fork of the [vitess project’s cgzip](https://github.com/youtube/vitess/tree/master/go/cgzip) package. Our primary data pipeline uses zlib compressed messages, but the standard library’s pure Go implementation can be significantly slower than the C zlib library. In order to address this gap, we modified a few flags in cgzip to make it encode and decode with zlib wrapping rather than with gzip headers.

We’ve detailed some of the other more novel design decisions in czlib, including its batch interfaces, in our [general blog on performance in Go](https://www.datadoghq.com/blog/go-performance-tales/) a couple of years ago. Performance varies quite a bit among the various interfaces, so it pays to benchmark using a message that is typical for your system by running the czlib benchmark suite with `PAYLOAD=path_to_message go test -run=NONE -bench .`

Here are some benchmark results for compression and decompression of czlib compared to the standard library:
```
go version go1.22.6 darwin/arm64
pkg: github.com/DataDog/czlib

# 2KiB file
     │ CompressStdZlib │               Compress               │
     │     sec/op      │    sec/op     vs base                │
*-10      75.20µ ± 12%   39.84µ ± 31%  -47.02% (p=0.000 n=10)
     │ CompressStdZlib │               Compress                │
     │       B/s       │      B/s       vs base                │
*-10     27.71Mi ± 11%   52.30Mi ± 24%  +88.73% (p=0.000 n=10)

     │ DecompressStdZlib │             Decompress              │
     │      sec/op       │   sec/op     vs base                │
*-10        18.353µ ± 5%   4.993µ ± 4%  -72.80% (p=0.000 n=10)
     │ DecompressStdZlib │              Decompress               │
     │        B/s        │     B/s       vs base                 │
*-10        113.5Mi ± 5%   417.4Mi ± 3%  +267.60% (p=0.000 n=10)

# Silesia compression corpus - mr (~10MB)
     │ CompressStdZlib │              Compress               │
     │     sec/op      │   sec/op     vs base                │
*-10       327.1m ± 1%   381.0m ± 1%  +16.46% (p=0.000 n=10)

     │ CompressStdZlib │               Compress               │
     │       B/s       │     B/s       vs base                │
*-10      29.07Mi ± 1%   24.96Mi ± 1%  -14.14% (p=0.000 n=10)

     │ DecompressStdZlib │             Decompress              │
     │      sec/op       │   sec/op     vs base                │
*-10         51.20m ± 1%   13.96m ± 2%  -72.74% (p=0.000 n=10)
     │ DecompressStdZlib │              Decompress               │
     │        B/s        │     B/s       vs base                 │
*-10        185.7Mi ± 1%   681.2Mi ± 2%  +266.81% (p=0.000 n=10)
```

[See more on the blog post](https://www.datadoghq.com/blog/engineering/releasing-czlib-and-zstd-go-bindings/)