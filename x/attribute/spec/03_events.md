# Events

The attribute module emits the following events:

<!-- TOC 2 2 -->
  - [Attribute Added](#attribute-added)
  - [Attribute Updated](#attribute-updated)
  - [Attribute Expiration Updated](#attribute-expiration-updated)
  - [Attribute Deleted](#attribute-deleted)
  - [Distinct Attribute Deleted](#distinct-attribute-deleted)
  - [Attribute Expired](#attribute-expired)
  - [Account Data Updated](#account-data-updated)

---
## Attribute Added

Fires when an attribute is successfully added.

| Type              | Attribute Key | Attribute Value        |
|-------------------|---------------|------------------------|
| EventAttributeAdd | Name          | \{name string\}          |
| EventAttributeAdd | Value         | \{attribute value\}      |
| EventAttributeAdd | Type          | \{attribute value type\} |
| EventAttributeAdd | Account       | \{account address\}      |
| EventAttributeAdd | Owner         | \{owner address\}        |
| EventAttributeAdd | Expiration    | \{expiration date/time\} |

`provenance.attribute.v1.EventAttributeAdd`

---
## Attribute Updated

Fires when an existing attribute is successfully updated.

| Type                 | Attribute Key | Attribute Value            |
|----------------------|---------------|----------------------------|
| EventAttributeUpdate | Name          | \{name string\}              |
| EventAttributeUpdate | OriginalValue | \{attribute value\}          |
| EventAttributeUpdate | OriginalType  | \{attribute value type\}     |
| EventAttributeUpdate | UpdateValue   | \{new attribute value\}      |
| EventAttributeUpdate | UpdateType    | \{new attribute value type\} |
| EventAttributeUpdate | Account       | \{account address\}          |
| EventAttributeUpdate | Owner         | \{owner address\}            |

`provenance.attribute.v1.EventAttributeUpdate`

---
## Attribute Expiration Updated

Fires when an existing attribute's expiration is successfully updated.

| Type                           | Attribute Key      | Attribute Value            |
|--------------------------------|--------------------|----------------------------|
| EventAttributeExpirationUpdate | Name               | \{name string\}              |
| EventAttributeExpirationUpdate | Value              | \{attribute value\}          |
| EventAttributeExpirationUpdate | Account            | \{account address\}          |
| EventAttributeExpirationUpdate | Owner              | \{owner address\}            |
| EventAttributeExpirationUpdate | OriginalExpiration | \{old expiration date/time\} |
| EventAttributeExpirationUpdate | UpdatedExpiration  | \{new expiration date/time\} |


---
## Attribute Deleted

Fires when an existing attribute is deleted.

| Type                 | Attribute Key | Attribute Value   |
|----------------------|---------------|-------------------|
| EventAttributeDelete | Name          | \{name string\}     |
| EventAttributeDelete | Account       | \{account address\} |
| EventAttributeDelete | Owner         | \{owner address\}   |

`provenance.attribute.v1.EventAttributeDelete`

---
## Distinct Attribute Deleted

Fires when an existing attribute is deleted distinctly.

| Type                         | Attribute Key | Attribute Value        |
|------------------------------|---------------|------------------------|
| EventAttributeDistinctDelete | Name          | \{name string\}          |
| EventAttributeDistinctDelete | Value         | \{attribute value\}      |
| EventAttributeDistinctDelete | AttributeType | \{attribute value type\} |
| EventAttributeDistinctDelete | Owner         | \{owner address\}        |
| EventAttributeDistinctDelete | Account       | \{account address\}      |

`provenance.attribute.v1.EventAttributeDistinctDelete`

---
## Attribute Expired

Fires when an attribute's expriration date/time has been reached and the attribute has been deleted.

| Type                  | Attribute Key | Attribute Value        |
|-----------------------|---------------|------------------------|
| EventAttributeExpired | Name          | \{name string\}          |
| EventAttributeExpired | Value         | \{attribute value\}      |
| EventAttributeExpired | AttributeType | \{attribute value type\} |
| EventAttributeExpired | Account       | \{account address\}      |
| EventAttributeExpired | Owner         | \{owner address\}        |
| EventAttributeExpired | Expiration    | \{expiration date/time\} |

---
## Account Data Updated

Fires when account data is updated for an account.

| Type                    | Attribute Key | Attribute Value        |
|-------------------------|---------------|------------------------|
| EventAccountDataUpdated | Account       | \{account address\}      |
