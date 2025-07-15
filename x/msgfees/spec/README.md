# `x/msgfees`

## Overview

The msgfees module has been deprecated and largely removed in favor of the `x/flatfees` module.

The msgfees module no longer stores anything in state and no longer has any `Msg` endpoints (governance or otherwise).
Its `Msg` types have been retained so that Txs involving them can still be read from state, but the endpoints for them no longer exist.

## Contents

1. **[Queries](04_queries.md)**
