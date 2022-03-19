# BGPiano

BGP message <-> MIDI message, for times when you want to broadcast your music instead of your IP packets.

## Usage

WIP.

## Building

### Windows Support

GoBGP does not support Windows natively. To build this project under Windows with a little hack, use the following
method:

1. Clone `https://github.com/osrg/gobgp.git` somewhere outside this directory
2. Apply `contrib\windows\gobgp-windows.patch` to the GoBGP source directory
3. Append `replace github.com/osrg/gobgp/v3 => ../relative/path/to/gobgp` to `go.mod` in BGPiano project directory
4. Build
