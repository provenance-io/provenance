# Events

The name module emits the following events:

<!-- TOC -->
  - [Handlers](#handlers)
    - [MsgBindNameRequest](#msgbindnamerequest)
    - [MsgDeleteNameRequest](#msgdeletenamerequest)
    - [MsgModifyNameRequest](#msgmodifynamerequest)
    - [CreateRootNameProposal](#createrootnameproposal)

## Handlers

### MsgBindNameRequest

| Type                  | Attribute Key         | Attribute Value           |
| --------------------- | --------------------- | ------------------------- |
| name_bound            | name                  | \{NameRecord|Name\}         |
| name_bound            | address               | \{NameRecord|Address\}      |
| name_bound            | restricted            | \{NameRecord|Restricted\}   |


### MsgDeleteNameRequest

| Type                  | Attribute Key         | Attribute Value           |
| --------------------- | --------------------- | ------------------------- |
| name_unbound          | name                  | \{NameRecord|Name\}         |
| name_unbound          | address               | \{NameRecord|Address\}      |
| name_unbound          | restricted            | \{NameRecord|Restricted\}   |

### MsgModifyNameRequest

| Type                  | Attribute Key         | Attribute Value           |
| --------------------- | --------------------- | ------------------------- |
| name_modify           | authority             | \{String\}                  |
| name_modify           | name                  | \{NameRecord|Name\}         |
| name_modify           | address               | \{NameRecord|Address\}      |
| name_modify           | restricted            | \{NameRecord|Restricted\}   |


### CreateRootNameProposal

| Type                  | Attribute Key         | Attribute Value           |
| --------------------- | --------------------- | ------------------------- |
| name_bound            | name                  | \{NameRecord|Name\}         |
| name_bound            | address               | \{NameRecord|Address\}      |
| name_bound            | restricted            | \{NameRecord|Restricted\}   |
