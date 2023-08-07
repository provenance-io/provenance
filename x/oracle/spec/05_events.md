<!--
order: 5
-->

# Events

The oracle module emits the following events:


---
## EventOracleQuerySuccess

This event is emitted when a `ICQ` response is received from an `ACK` and is successful.

| Type               | Attribute Key | Attribute Value                 |
| ------------------ | ------------- | ------------------------------- |
| OracleQuerySuccess | sequence_id   | Sequence ID of the ICQ request  |
| OracleQuerySuccess | result        | Query data obtained from oracle |

---
## EventOracleQueryError

This event is emitted when a `ICQ` response is received from an `ACK` and contains an error.

| Type             | Attribute Key | Attribute Value                |
| ---------------- | ------------- | ------------------------------ |
| OracleQueryError | sequence_id   | Sequence ID of the ICQ request |
| OracleQueryError | error         | Error received from the module |

---
## EventOracleQueryTimeout

This event is emitted when a `ICQ` request results in a `Timeout`.

| Type               | Attribute Key | Attribute Value                |
| ------------------ | ------------- | ------------------------------ |
| OracleQueryTimeout | sequence_id   | Sequence ID of the ICQ request |