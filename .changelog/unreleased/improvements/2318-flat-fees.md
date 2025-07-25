* Triggers no longer track or use the extra gas provided when creating the trigger [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
  Users pay for the trigger msg execution when creating the trigger (based on msg type).
* Renamed `pioconfig` stuff to use `Prov` instead of `Provenance` [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
* Make `pioconfig.GetProvConfig` return the defaults if `SetProvConfig` hasn't been called yet [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
