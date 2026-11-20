* Marker accounts returned from the auth module queries will always have an empty access list [#2771](https://github.com/provenance-io/provenance/issues/2771).
  Use the marker module queries to get access list information.
* The `MarkerAccountI` methods that get or modify the access list have been deprecated and should not longer be used [#2771](https://github.com/provenance-io/provenance/issues/2771).
  Keeper methods should be used instead.
