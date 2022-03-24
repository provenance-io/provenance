
### Summary

Provenance 1.8.0 is focused on improving the fee structures for transactions on the blockchain.  While the Cosmos SDK has traditionally offered a generic fee structure focused on gas/resource utilization, the Provenance blockchain has found that certain transactions have additional long term costs and value beyond simple resources charges.  This is the reason we are adding the new MsgFee module which allows governance based control of additional fee charges on certain message types.

NOTE: The second major change in the 1.8.0 release is part of the migration process which removes many orphaned state objects that were left in 1.7.x chains.  This cleanup process will require a significant amount of time to perform during the `green` upgrade handler execution.  The upgrade will print status messages showing the progress of this process.



# Performing the 1.8.0 Upgrade (blockheight 4808400)

The `1.8.0` upgrade process will begin with the current 1.7.x binary halting at the assigned upgrade height (4808400). The log will print "UPGRADE NEEDED".

1. Terminate the existing 1.7.x process after the upgrade height is reached.
2. Create a backup of your local data directory.  This will allow you to quickly reset to the state prior to the upgrade changes in the event that the upgrade fails and the network needs to 'skip the upgrade' for some reason.
3.  Copy the new `1.8.0` binary and lib wasm shared library into your provenance home folder.  It can be helpful to name your binaries with the version associated with them (`provenanced-1.8.0` for example).
  NOTE: the libwasmvm.so share object library must match the version of the provenance binary used.  If you do not match these versions correctly your node will halt with an APP-HASH-MISMATCH error as soon as it receives a smart contract request.
4.  Ensure your `wasm/cache` folder has been cleared out.  `data/wasm/wasm/cache`.  These cache files are not compatible between software versions.
5. If you are using a log level of `error` then it would be a good time to switch to `info` for this upgrade. The `--log_level=info` can be appended to your start command to temporarily change this setting. Start the 1.8.0 provenanced binary.  The upgrade should begin shortly after it starts.  There will be various errors printed due to peering issues during the upgrade ... this is normal and not something to worry about.
6. The upgrade may take 30 minutes or so depending on your hardware.  The following set of INFO level log messages are expected.  NOTE: The actual counts will vary slightly depending on the number of new database entries created in the last couple days between when this test was ran and the upgrade height.
```
3:36PM INF Updated governance module minimum deposit from 1000000000000nhash to 50000000000000nhash
3:36PM INF Updated attribute module max length value from 10000 to 1000
3:36PM INF NOTICE: Starting migrations. This may take a significant amount of time to complete. Do not restart node.
3:36PM INF Migrating Metadata Module from Version 2 to 3
3:36PM INF Metadata step 1 of 7: Deleting bad indexes and getting good
3:38PM INF Prefix 17: Done. Deleted 104354 entries, found 202006 good entries.
3:38PM INF Prefix 18: Identifying good entries and deleting bad ones.
3:39PM INF Prefix 18: Done. Deleted 52808 entries, found 101714 good entries.
3:39PM INF Prefix 19: Identifying good entries and deleting bad ones.
3:39PM INF Prefix 19: Done. Deleted 4 entries, found 6 good entries.
3:39PM INF Prefix 20: Identifying good entries and deleting bad ones.
3:39PM INF Prefix 20: Done. Deleted 2607 entries, found 3173 good entries.
3:39PM INF Done identifying good indexes and deleting bad ones.
3:39PM INF Metadata step 2 of 7: Reindexing scopes
3:40PM INF Done reindexing 10449 scopes. All 112164 are now indexed.
3:40PM INF Metadata step 3 of 7: Reindexing scope specs
3:40PM INF Done reindexing 0 scope specs. All 6 are now indexed.
3:40PM INF Metadata step 4 of 7: Reindexing contract specs
3:40PM INF Done reindexing 2607 contract specs. All 5780 are now indexed.
3:40PM INF Metadata step 5 of 7: Identifying sessions to keep
3:48PM INF Done identifying a total of 237678 session ids used by 1171973 records.
3:48PM INF Metadata step 6 of 7: Finding empty sessions
3:55PM INF Done deleting 640843 empty sessions.
3:55PM INF Finished Migrating Metadata Module from Version 2 to 3
3:55PM INF Adding a 10 hash (10,000,000,000 nhash) msg fee to MsgBindNameRequest (1/6)
3:55PM INF Adding a 100 hash (100,000,000,000 nhash) msg fee to MsgAddMarkerRequest (2/6)
3:55PM INF Adding a 10 hash (10,000,000,000 nhash) msg fee to MsgAddAttributeRequest (3/6)
3:55PM INF Adding a 10 hash (10,000,000,000 nhash) msg fee to MsgWriteScopeRequest (4/6)
3:55PM INF Adding a 10 hash (10,000,000,000 nhash) msg fee to MsgP8EMemorializeContractRequest (5/6)
3:55PM INF Adding a 100 hash (100,000,000,000 nhash) msg fee to MsgSubmitProposal (6/6)
3:55PM INF Successfully upgraded to: green with version map: map[attribute:2 auth:2 authz:1 bank:2 capability:1 crisis:1 distribution:2 evidence:1 feegrant:1 genutil:1 gov:2 ibc:2 marker:2 metadata:3 mint:1 msgfees:1 name:2 params:1 slashing:2 staking:2 transfer:1 upgrade:1 vesting:1 wasm:1]
```
7. When the upgrade has finished your node will wait for other peers and the validators to complete the upgrade as well.  The active voting power must reach the 2/3 majority threshold for the network to continue.


# What to do if something goes wrong

First off, make sure you have contacted the other network participants on the Provenance discord server. https://discord.gg/HszDgS6R

The network can proceed one of two ways, ideally the majority of nodes are able to execute the 1.8.0 migration and the network continues on 1.8.0.  If your node fails to complete these steps a quicksync file could be used to skip over the upgrade process using someone else's completed data files.

## Skipping the Upgrade
If the upgrade is broken and fails completely a decision may be made to skip the upgrade entirely.  If this happens there will be a bunch of discussion and announcements on the discord channel for it.  In this case you would use the backup file taken before the upgrade, the previous `1.7.x` binary/wasm vm dll along with the `--unsafe-skip-upgrades 4808400` flag.  This option will only work if everyone is using this same flag... if you attempt this independently it will cause your node to compute an incorrect app hash resulting in a corrupted data directory that you must restore from backup.

## Worst Case Recovery
And the absolute worst case scenario would be a network that proceeds few blocks post upgrade then halts due to some undiscovered issue.  While we have extensively tested this release (including 10 different successful executions of the upgrade on testnet as well as isolated tests against mainnet data) it is possible that a bug exists that halts the network.  If this happens the only fix will be an emergency software release with patches to address the issue which everyone will need to install and use.  If this situation occurs there will be extensive coordination via the Discord channels.  The network would likely be halted for a number of hours prior to being able to attempt to restart it.

