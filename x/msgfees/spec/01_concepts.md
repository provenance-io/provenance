<!--
order: 1
-->

# Concepts

## Additional Msg Fees

Fees is one of the most important tools available to secure a PoS network since incentivizes staking and encourages spam prevention.
For e.g a Scope created on the Provenance network is way more valuable as an Asset that has real world value, much greater that a name or attribute 
assigned to an address.


## Additional fee collection Procedure

1. ante handler gets called which calls the FeeDecorator

2. FeeDecorator calls MsgFee module keeper and figures out the additional fees.

3. On simulation calculated fee is simply returned.

4. On actual Tx, check account to see if they can cover the fees, if they can send 
fees to Fee module for distribution(at least i think that is what shpuld happen)

5. Add gov proposals for msg's to be charged extra fee's

6.Does wasm work?
