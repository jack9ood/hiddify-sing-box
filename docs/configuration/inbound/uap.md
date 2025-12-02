### Structure

```json
{
  "type": "uap",
  "tag": "uap-in",

  ... // Listen Fields

  "users": [
    {
      "name": "sekai",
      "uuid": "bf000d23-0752-40b4-affe-68f7707a9661",
      "flow": ""
    }
  ],
  "tls": {},
  "multiplex": {},
  "transport": {}
}
```

### Listen Fields

See [Listen Fields](/configuration/shared/listen/) for details.

### Fields

#### users

==Required==

UAP users.

#### users.uuid

==Required==

UAP user id.

#### users.flow

UAP Sub-protocol.

Available values:

* `xtls-rprx-vision`

#### tls

TLS configuration, see [TLS](/configuration/shared/tls/#inbound).

!!! note "Reality Support"
    UAP fully supports Reality obfuscation when configured with `flow: "xtls-rprx-vision"` and appropriate TLS/Reality settings.

#### multiplex

See [Multiplex](/configuration/shared/multiplex#inbound) for details.

#### transport

V2Ray Transport configuration, see [V2Ray Transport](/configuration/shared/v2ray-transport/).

### Protocol Details

!!! info "UAP Protocol"
    UAP is a custom protocol based on VLESS but with the following differences:
    
    - **Protocol Version**: 1 (incompatible with VLESS version 0)
    - **Wire Format**: Optimized binary format without Protobuf dependency
    - **Performance**: Reduced serialization overhead for improved throughput
    - **Reality**: Full support for Reality obfuscation with Vision flow control
