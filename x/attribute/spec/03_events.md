# Events

The attribute module emits the following events:

<!-- TOC 2 2 -->
  - [Attribute Added](#attribute-added)
  - [Attribute Updated](#attribute-updated)
  - [Attribute Deleted](#attribute-deleted)
  - [Distinct Attribute Deleted](#distinct-attribute-deleted)

---
## Attribute Added

Fires when an attribute is successfully added.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| EventAttributeAdd      | Name                  | {name string}             |
| EventAttributeAdd      | Value                 | {attribute value}         |
| EventAttributeAdd      | Type                  | {attribute value type}    |
| EventAttributeAdd      | Account               | {account address}         |
| EventAttributeAdd      | Owner                 | {owner address}           |

`provenance.attribute.v1.EventAttributeAdd`

---
## Attribute Updated

Fires when an existing attribute is successfully updated.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| EventAttributeUpdate   | Name                  | {name string}             |
| EventAttributeUpdate   | OriginalValue         | {attribute value}         |
| EventAttributeUpdate   | OriginalType          | {attribute value type}    |
| EventAttributeUpdate   | UpdateValue           | {new attribute value}     |
| EventAttributeUpdate   | UpdateType            | {new attribute value type}|
| EventAttributeUpdate   | Account               | {account address}         |
| EventAttributeUpdate   | Owner                 | {owner address}           |

`provenance.attribute.v1.EventAttributeUpdate`

---
## Attribute Deleted

Fires when an existing attribute is deleted.

| Type                     | Attribute Key         | Attribute Value           |
| ------------------------ | --------------------- | ------------------------- |
| EventAttributeDelete     | Name                  | {name string}             |
| EventAttributeDelete     | Account               | {account address}         |
| EventAttributeDelete     | Owner                 | {owner address}           |

`provenance.attribute.v1.EventAttributeDelete`

---
## Distinct Attribute Deleted


Fires when an existing attribute is deleted distinctly.

| Type                          | Attribute Key         | Attribute Value           |
| ----------------------------- | --------------------- | ------------------------- |
| EventAttributeDistinctDelete  | Name                  | {name string}             |
| EventAttributeDistinctDelete  | Value                 | {attribute value}         |
| EventAttributeDistinctDelete  | Owner                 | {owner address}           |
| EventAttributeDistinctDelete  | Account               | {account address}         |

`provenance.attribute.v1.EventAttributeDistinctDelete`

---