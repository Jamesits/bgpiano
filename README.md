# BGPiano

BGP message <-> MIDI message, for times when you want to broadcast your music instead of your IP packets.

## Usage

### Point to Point

The GoBGP library we use does not support customizing peer TCP port. Thus, you are stuck with port 179 and would
(in most cases) need root privilege to listen on that port.

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

Go 1.18 or higher is required.

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

- Check PIE: `checksec --dir=dist` (should be all green)

### Windows

GoBGP does not support Windows officially. But don't worry! We understand music production is hard under Linux and you
might have connected all your instruments to your Windows computer. This project equally supports Windows. (Actually,
most of the development work is done under Windows.)

The only thing you need is a little hack on the GoBGP library:

```powershell
cd ..
git clone https://github.com/osrg/gobgp.git
cd gobgp
git apply ..\bgpiano\contrib\windows\gobgp-windows.patch
cd ..\bgpiano
cp contrib\windows\_go.work go.work
```

After this, build the application you need with `go build ./cmd/<executable_name>`.

## FAQ

### Why?

There is a trend that Chinese BGP players[^bgp-players] misuse 广播 (*lit.* broadcasting) in the meaning of 宣告 (*lit.*
announcement). As a terminology fundamentalist and one of the earliest BGP players, I hate this incorrect usage of word
to the bone. However, this wrong terminology is now widespread, so I decided to fix it the other way around, by
literally *broadcasting* a piece of music across the Internet, through the BGP RIB.

### How

The MIDI message is encoded in either extended community or large community. See [protocol.md](doc/protocol.md) for
details.

### Will this add additional pressure to the routers?

Yes, but unless you programmatically play something violent like
[最終鬼畜妹フランドール・S](https://www.youtube.com/watch?v=ql-Rvn50p-Y), a modern router should handle it pretty easily.

### What's the latency?

For BGP propagation latency across the globe,
*[The speed of BGP network propagation by Ben Cox](https://blog.benjojo.co.uk/post/speed-of-bgp-network-propagation)*
already provided an excellent overview.

On the topic of latency deviation, due to how BGP routes are received and updated, there are no guarantee that any
individual note will arrive on time. [Why not have some Jazz instead?](https://www.youtube.com/watch?v=lpc1lEJ-SRc)

<!-- footnotes -->

[^bgp-players]: slang for public ASN owners who use the ASN only for education, research and
[zhuangbility](https://www.urbandictionary.com/define.php?term=zhuangbility).