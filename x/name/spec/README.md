# `name`

## Overview

The name service is intended to provide a system for creating human-readable names as aliases for addresses and to imply ownership and control.  These names can be used to provide a stable reference to a changing address or collection of addresses.

One issue with a blockchain is that addresses are complex strings of characters that are difficult to type and remember.  On the other hand the name service can provide a potentially shorter and easier to remember alias such as `provenance.pb` or `attribute.user.pb` to use in place of the address.

### A Name Hierarchy

Another challenge for users of a blockchain is establishing authority and delegating control.  A specific example of this is the definition of the authoritative source of a piece of information.  Where did this information come from? Who created it/vetted it?  How can this control be distributed in such a way that the right people can control the information?  A narrow aspect of this type of control can be satisfied through the creation of a hierarchical name system modeled after DNS.  If the address `passport.pb` has been created and is owned by the Provenance Passport application, then `level-3.accredited.passport.pb` can be expected to be under the direct or delegated control of the passport application.

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
    - [MsgBindNameRequest](03_messages.md#msgbindnamerequest)
    - [MsgDeleteNameRequest](03_messages.md#msgdeletenamerequest)
    - [CreateRootNameProposal](03_messages.md#createrootnameproposal))
4. **[Events](04_events.md)**
    - [Handlers](04_events.md#handlers)
7. **[Parameters](05_params.md)**