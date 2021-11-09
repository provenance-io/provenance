# Metadata Authz

The `authz` implementation in the `metadata` module checks for granted permission in cases when there are missing signatures.

A `GenericAuthorization` should be used using the message type URLs now documented in [03_messages.md](03_messages.md).

<!-- TOC -->
  - [Code](#code)
  - [CLI](#cli)
  - [Special allowances](#special-allowances)

---

## Code

Grant:
```aspectj
granter := ... // Bech32 AccAddress
grantee := ... // Bech32 AccAddress
a := authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeRequest)
err := s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, now.Add(time.Hour))
```

Delete:
```aspectj
err := s.app.AuthzKeeper.DeleteGrant(s.ctx, grantee, granter, types.TypeURLMsgWriteScopeRequest)
```
Revoke:
```aspectj
granter := ... // Bech32 AccAddress
grantee := ... // Bech32 AccAddress
msgRevoke := authz.NewMsgRevoke(granter, grantee, types.TypeURLMsgWriteScopeRequest)
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


See [GenericAuthorization](https://docs.cosmos.network/master/architecture/adr-030-authz-module.html#genericauthorization) specification for more details.

## Special allowances

Some messages in the `metadata` module have hierarchies. A grant on a parent message type will also work for any of 
its message subtypes, but not the other way around. Therefore, authorizations on these messages are `one way`. 

- An authorization on `MsgWriteScopeRequest` works for any of the listed message subtypes:
  - `MsgAddScopeDataAccessRequest`
  - `MsgAddScopeDataAccessRequest`
  - `MsgDeleteScopeDataAccessRequest` 
  - `MsgAddScopeOwnerRequest`
  - `MsgDeleteScopeOwnerRequest`

- An authorization on `MsgWriteSessionRequest` works for any of the listed message subtypes:
    - `MsgWriteRecordRequest`

- An authorization on `MsgWriteScopeSpecificationRequest` works for any of the listed message subtypes:
    - `MsgAddContractSpecToScopeSpecRequest`
    - `MsgDeleteContractSpecFromScopeSpecRequest`

- An authorization on `MsgWriteContractSpecificationRequest` works for any of the listed message subtypes:
    - `MsgWriteRecordSpecificationRequest`

- An authorization on `MsgDeleteContractSpecificationRequest` works for any of the listed message subtypes:
    - `MsgDeleteRecordSpecificationRequest`
    

Notes: 

An authorization on a `Write` endpoint for an entry/spec will NOT work for its `Delete` endpoint. 
