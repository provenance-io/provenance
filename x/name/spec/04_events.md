# Events

The name module emits the following events:

## Handlers

### MsgBindNameRequest

| Type                  | Attribute Key         | Attribute Value           |
| --------------------- | --------------------- | ------------------------- |
| name_bound            | name                  | {NameRecord|Name}         |
| name_bound            | address               | {NameRecord|Address}      |
| name_bound            | restricted            | {NameRecord|Restricted}   |


### MsgDeleteNameRequest

| Type                  | Attribute Key         | Attribute Value           |
| --------------------- | --------------------- | ------------------------- |
| name_unbound          | name                  | {NameRecord|Name}         |
| name_unbound          | address               | {NameRecord|Address}      |
| name_unbound          | restricted            | {NameRecord|Restricted}   |


### CreateRootNameProposal

| Type                  | Attribute Key         | Attribute Value           |
| --------------------- | --------------------- | ------------------------- |
| name_bound            | name                  | {NameRecord|Name}         |
| name_bound            | address               | {NameRecord|Address}      |
| name_bound            | restricted            | {NameRecord|Restricted}   |


### ModifyNameProposal

| Type                  | Attribute Key         | Attribute Value           |
| --------------------- | --------------------- | ------------------------- |
| name_update           | name                  | {NameRecord|Name}         |
| name_update           | address               | {NameRecord|Address}      |
| name_update           | restricted            | {NameRecord|Restricted}   |