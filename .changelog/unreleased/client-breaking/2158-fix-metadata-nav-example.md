* Fixes the metadata nav cli command example to use the correct module name [#2058](https://github.com/provenance-io/provenance/issues/2058).
  During this fix it was discovered that the volume parameter was not present but was required for proper price ratios.  The volume
  parameter has been added to the NAV entry and when not present a default value of 1 (which should be the most common case for a scope) is
  used instead.
