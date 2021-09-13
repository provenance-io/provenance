## Development Docker Instance

# Overview

This folder contains the files to create an single node with predefined account stocked with nhash.

# Adding Account

Currently, in the folder `networks/dev/mnemonics` there exists 3 files containing a mnemonic and a single return character.  These are used to generate an account into the genesis file with `100000000000000000000nhash` each.  A new key will be added to the keyring that is the basename of the file.  Therefore, `validator.txt`, `account-1.txt`, and `account-2.txt` will add keys `validator`, `account-1` and `account-2`.  The `validator.txt` is required, because it is used to setup the validator and is nhash marker admin.  Other files can be added and removed as you see fit.

# Added Account Addresses and Keys 

```
- name: account-1
  type: local
  address: tp1uncq4lmkffy2crtxpuw755946c5wray2uw42qx
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Aln9OfLgmV4mhKHKIeFfY1nuXEs0VAG4anwXdnDr64Lw"}'
  mnemonic: ""

- name: account-2
  type: local
  address: tp1l2f7duvp36j7mc0vfwevcp5apaed9682hrtng5
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Ajhmuy4AnN1g1DLMHv/r5kjINrW84vOFXSvQfY8gxUDD"}'
  mnemonic: ""

- name: validator
  type: local
  address: tp1yvpv5yv4jcuckxk3ud2evskmpt0ks8ej9ectxy
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AmNiOomRpLVkzwGRWfhHJJv1RFycKX9gkGZyvoplx6F3"}'
  mnemonic: ""
```


NOTE: This folder contains files to run an image using a locally built binary 
that leverages Go's ability to target platforms during builds.  These docker
files are _not_ the ones used to build the release image.  For those images
see the `docker/blockchain` folder in the project root.

