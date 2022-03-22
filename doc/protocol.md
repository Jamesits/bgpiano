# BGPiano Protocol

## Signaling

### Note On / Note Off

To comply with the metaphor of a key being pressed and released, note on / off messages should be mapped to route add /
withdraw messages. The route itself doesn't matter and should be ignored; the event should be encoded in an extended
community in the following format:

- Type: 0x88 (1 byte)[^RFC4360][^1][^RFC7153]
- Subtype: 0x00 (1 byte)
- Value (6 bytes)
    - channel (1 byte)
    - note (1 byte)
    - velocity (1 byte)
    - undefined (3 byte)

### General MIDI Message

### Extended Community

- Type: 0x88 (1 byte)[^RFC4360][^1][^RFC7153]
- Length (1 byte)
- MIDI message (max. 6 bytes)

### Large Community

TBD.

<!-- references -->
[^1]: [Border Gateway Protocol (BGP) Extended Communities - IANA](https://www.iana.org/assignments/bgp-extended-communities/bgp-extended-communities.xhtml)
[^RFC7153]: [IANA Registries for BGP Extended Communities](https://www.rfc-editor.org/rfc/rfc7153.html)
[^RFC4360]: [BGP Extended Communities Attribute](https://datatracker.ietf.org/doc/html/rfc4360)