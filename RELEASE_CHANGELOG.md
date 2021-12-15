## [v1.7.6](https://github.com/provenance-io/provenance/releases/tag/v1.7.6) - 2021-12-15

### Improvements

* Upgrade Rosetta to v0.7.2 [#560](https://github.com/provenance-io/provenance/issues/560)
    * Note: This update is only for Rosetta.  If you are not running a rosetta server this update will not change anything.
    * This update brings in a fix to Rosetta allow it to run against chains with coins minted post genesis.  In order to do this we updated the cosmos-sdk dependency to point to the provenance fork of v0.44.3 whose only change is having the most recent version of Rosetta, v0.7.2.