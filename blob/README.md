# blob

A library for loading configuration files into [config](https://github.com/warthog618/config/tree/master).

[![GoDoc](https://godoc.org/github.com/warthog618/config/blob/sar?status.svg)](https://godoc.org/github.com/warthog618/config/blob)

Blobs represent configuration sources containing configuration stored as a block
and in a known format, e.g. config files.  Blobs are partitioned into two
layers, the Loader, which loads the configuration as a []byte blob from the
source, and the Decoder, which decodes the blob into a form that can be used by
config.

The Loader may support being watched for changes, and if so the Blob
containing it will automatically support being watched for changes as well.

## Quick Start

Given a loader and decoder, you create a blob and add it to your config:

```go
    c := config.NewConfig(blob.New(file.New("config.json"), json.NewDecoder()))
```

or add it into your config stack:

```go
    sources := config.NewStack(pflag.New(),env.New())
    c := config.NewConfig(sources)
    cfgFile := blob.New(file.New("config.json"), json.NewDecoder())
    sources.Append(cfgFile)
```

## Loaders

Loaders read configuration from some source.

The following loaders are provided:

Loader | Configuration Source
:-----:| -----
[bytes](https://github.com/warthog618/config/tree/master/blob/loader/bytes) | []byte
[file](https://github.com/warthog618/config/tree/master/blob/loader/file) | local file

## Decoders

Decoders unmarshal configuration from a particular textual format.

Decoders for the following formats are provided:

- [JSON](https://github.com/warthog618/config/tree/master/blob/decoder/json)
- [TOML](https://github.com/warthog618/config/tree/master/blob/decoder/toml)
- [YAML](https://github.com/warthog618/config/tree/master/blob/decoder/yaml)
- [HCL](https://github.com/warthog618/config/tree/master/blob/decoder/hcl)
- [INI](https://github.com/warthog618/config/tree/master/blob/decoder/ini)
- [properties](https://github.com/warthog618/config/tree/master/blob/decoder/properties)
