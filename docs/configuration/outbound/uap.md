### Structure

```json
{
  "type": "uap",
  "tag": "uap-out",
  
  "server": "127.0.0.1",
  "server_port": 1080,
  "uuid": "bf000d23-0752-40b4-affe-68f7707a9661",
  "flow": "xtls-rprx-vision",
  "network": "tcp",
  "tls": {},
  "packet_encoding": "",
  "multiplex": {},
  "transport": {},

  ... // Dial Fields
}
```

### Fields

#### server

==Required==

The server address.

#### server_port

==Required==

The server port.

#### uuid

==Required==

UAP user id.

#### flow

UAP Sub-protocol.

Available values:

* `xtls-rprx-vision`

#### network

Enabled network

One of `tcp` `udp`.

Both is enabled by default.

#### tls

TLS configuration, see [TLS](/configuration/shared/tls/#outbound).

!!! note "Reality Support"
    UAP fully supports Reality obfuscation when configured with `flow: "xtls-rprx-vision"` and appropriate TLS/Reality settings.

#### packet_encoding

UDP packet encoding, xudp is used by default.

| Encoding   | Description           |
| ---------- | --------------------- |
| (none)     | Disabled              |
| packetaddr | Supported by v2ray 5+ |
| xudp       | Supported by xray     |

#### multiplex

See [Multiplex](/configuration/shared/multiplex#outbound) for details.

#### transport

V2Ray Transport configuration, see [V2Ray Transport](/configuration/shared/v2ray-transport/).

### Dial Fields

See [Dial Fields](/configuration/shared/dial/) for details.

### Protocol Details

!!! info "UAP Protocol"
    UAP is a custom protocol based on VLESS but with the following differences:
    
    - **Protocol Version**: 1 (incompatible with VLESS version 0)
    - **Wire Format**: Optimized binary format without Protobuf dependency
    - **Performance**: Reduced serialization overhead for improved throughput
    - **Reality**: Full support for Reality obfuscation with Vision flow control
    - **Compatibility**: UAP clients cannot connect to VLESS servers and vice versa
