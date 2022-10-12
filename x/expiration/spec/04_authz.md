# Metadata Authz

The `authz` implementation in the `expiration` module checks for granted permission in cases when there are missing signatures.

A `GenericAuthorization` will need to use the message type URLs documented in [03_messages.md](03_messages.md).

<!-- TOC -->
  - [Code](#code)
  - [CLI](#cli)

---

## Code

Grant:
```aspectj
granter := ... // Bech32 AccAddress
grantee := ... // Bech32 AccAddress
msg := types.MsgExtendExpirationRequest{}
a := authz.NewGenericAuthorization(msg.MsgTypeURL())
err := s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, now.Add(time.Hour))
```

Delete:
```aspectj
msg := types.MsgExtendExpirationRequest{}
err := s.app.AuthzKeeper.DeleteGrant(s.ctx, grantee, granter, msg.MsgTypeURL())
```
Revoke:
```aspectj
granter := ... // Bech32 AccAddress
grantee := ... // Bech32 AccAddress
msg := types.MsgExtendExpirationRequest{}
msgRevoke := authz.NewMsgRevoke(granter, grantee, msg.MsgTypeURL())
res, err := s.app.AuthzKeeper.Revoke(s.ctx, msgRevoke)
```

## CLI

Grant:
```aspectj
provenanced tx grant <grantee> <authorization> --from <granter>
```

Revoke:
```aspectj
provenanced tx revoke <grantee> <method-name> --from <granter>
```

See [GenericAuthorization](https://docs.cosmos.network/main/architecture/adr-030-authz-module.html#genericauthorization) specification for more details.
