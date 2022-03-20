# BGPiano

BGP message <-> MIDI message, for times when you want to broadcast your music instead of your IP packets.

## Usage

### Point to Point

Instrument (MIDI sender) side:

```shell
sudo bgpiano-send --bgp-port=179 --bgp-peer-ip=<peer-ip>
```

Synthesizer (MIDI receiver) side:

```shell
sudo bgpiano-recv --bgp-port=179 --bgp-peer-ip=<peer-ip>
```

### Reflected

Reflector side: `gobgp` or equivalent software required. Any RFC-compliant BGP daemon configured as an RR or RS can be
used.

```shell
sudo gobgpd --log-plain --config-file=contrib/rr-gobgp/gobgpd.toml
```

Instrument (MIDI sender) side:

```shell
bgpiano-send --bgp-peer-ip=<reflector-ip>
```

Synthesizer (MIDI receiver) side:

```shell
bgpiano-recv --bgp-peer-ip=<reflector-ip>
```

## Building

## Linux

Requirements:

- `$GOPATH` environment variable is set
- `$GOBIN` is in `$PATH`
- GCC is installed (for CGO)
- Dependencies: `libasound2-dev`

```shell
go install github.com/goreleaser/goreleaser@latest
goreleaser build --snapshot --rm-dist
```

Notes:
- Check PIE: `checksec --dir=dist`

### Windows Support

GoBGP does not support Windows natively. To build this project under Windows with a little hack, use the following
method:

1. Clone `https://github.com/osrg/gobgp.git` somewhere outside this directory
2. Apply `contrib\windows\gobgp-windows.patch` to the GoBGP source directory
3. Append `replace github.com/osrg/gobgp/v3 => ../relative/path/to/gobgp` to `go.mod` in BGPiano project directory
4. Build the application you need with `go build ./cmd/<executable_name>`
