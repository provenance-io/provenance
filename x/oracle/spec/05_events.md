<!--
order: 5
-->

# Events

The oracle module emits the following events:

<!-- TOC 2 -->
  - [EventOracleQuerySuccess](#eventoraclequerysuccess)
  - [EventOracleQueryError](#eventoraclequeryerror)
  - [EventOracleQueryTimeout](#eventoraclequerytimeout)


---
## EventOracleQuerySuccess

This event is emitted when an `ICQ` response is received from an `ACK` and is successful.

| Type               | Attribute Key | Attribute Value                     |
| ------------------ | ------------- | ----------------------------------- |
| OracleQuerySuccess | channel       | Channel the ICQ request was sent on |
| OracleQuerySuccess | sequence_id   | Sequence ID of the ICQ request      |
| OracleQuerySuccess | result        | Query data obtained from oracle     |

---
## EventOracleQueryError

This event is emitted when an `ICQ` response is received from an `ACK` and contains an error.

| Type             | Attribute Key | Attribute Value                     |
| ---------------- | ------------- | ----------------------------------- |
| OracleQueryError | channel       | Channel the ICQ request was sent on |
| OracleQueryError | sequence_id   | Sequence ID of the ICQ request      |
| OracleQueryError | error         | Error received from the module      |

---
## EventOracleQueryTimeout

This event is emitted when an `ICQ` request results in a `Timeout`.

| Type               | Attribute Key | Attribute Value                     |
| ------------------ | ------------- | ----------------------------------- |
| OracleQueryTimeout | channel       | Channel the ICQ request was sent on |
| OracleQueryTimeout | sequence_id   | Sequence ID of the ICQ request      |
