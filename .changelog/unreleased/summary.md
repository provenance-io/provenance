Provenance Blockchain version `v1.32.0` contains some exciting new features, improvements and bug fixes.

> [!IMPORTANT]
> **Advance notice: Linux release binaries will require a newer glibc in an upcoming release**
>
> This doesn't affect `v1.32.0`. In a future release, we'll move the Linux release build from Ubuntu 20.04 to Ubuntu 22.04.
> After that change, the `provenanced` and `libwasmvm` binaries in the Linux release zip will require **glibc 2.35 or newer**.
>
> To prepare, validators and node operators who use the published Linux release binaries should check their hosts now:
>
> ```bash
> ldd --version
> ```
>
> If it reports a version older than 2.35, plan to upgrade the operating system (e.g. to Ubuntu 22.04+ or Debian 12+) before that release.
> Hosts on Ubuntu 20.04, Debian 11, or RHEL/Rocky/Alma 8 will be affected.
> On a host that hasn't been upgraded, the new binary will fail to start with an error like ``version `GLIBC_2.35' not found``.
> This is especially important if you use Cosmovisor with auto-download, because the failure would happen at the upgrade height.
>
> Nodes built from source, or run from the official Docker images, are not affected.
