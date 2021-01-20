# Events

The name module emits the following events:

## Handlers

### MsgBindNameRequest

| Type                  | Attribute Key         | Attribute Value           |
| --------------------- | --------------------- | ------------------------- |
| name_bound            | name                  | {NameRecord|Name}         |
| name_bound            | address               | {NameRecord|Address}      |


### MsgDeleteNameRequest

| Type                  | Attribute Key         | Attribute Value           |
| --------------------- | --------------------- | ------------------------- |
| name_unbound          | name                  | {NameRecord|Name}         |
| name_unbound          | address               | {NameRecord|Address}      |
