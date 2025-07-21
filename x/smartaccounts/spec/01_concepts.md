<!--
order: 1
-->

# Concepts

## Limitations of Traditional Crypto Wallets

Externally owned accounts (EOAs), or traditional crypto wallets, come with several drawbacks:

- Single point of failure: One private key controls the entire account
    - High risk: Losing your private key means losing access to your funds permanently
    - Limited flexibility: Manual transfer of private key data between devices is required
    - Security concerns: No ability to rotate keys for enhanced protection
    - No key rotation possible

These constraints pose significant challenges, especially for newcomers to the crypto space. Users must quickly learn complex security practices to safeguard their assets effectively.

EOA-based systems typically lack robust multi-device support. Users who need to access their accounts on different devices often compromise their security by manually transferring recovery phrases.

Moreover, managing multiple networks, applications, and browser extensions not only adds complexity but also exposes users to security risks such as malware and phishing attacks.

The result is a complex and unforgiving user experience that falls short of the standards set by many modern digital services outside the crypto sphere.

# What are Smart accounts supposed to do

Smart Accounts allow transactions to be approved by multiple authentication methods rather than having to sign every transaction using a private key.

- Support Web 2.0 based standards like FIDO2.

*you can use familiar authentication technologies like Face ID, Touch ID, and Passkeys to easily create a new account, or manage or access existing accounts.*

- Support session based tokens so that we can support 1-click authentication for a short period of time (Sign a payload and create a session for a short time, this will greatly help user experience for trade placement imo)
- You are no longer tied to just a mnemonic.
- Traditional EOAs are still supported.
- Ability to re-register your new FIDO2 device, based on a certain flow (like use your old FIDO2 credentials or  use an EOA.(or maybe even some recovery way via an admin if they have opted in to it, this is more of a nice to have)


# Smart Accounts
Smart account extends the authentication capabilities of a base account.
It assumes authentication methods added to base account, these authentication methods can be of type
* Webauthn based
* k256 based
All of the authentication methods can be rotated if needed.

## Limitations of Traditional Crypto Wallets

Externally owned accounts (EOAs), or traditional crypto account, come with several drawbacks:

- Single point of failure: One private key controls the entire account
    - High risk: Losing your private key means losing access to your funds permanently
    - Limited flexibility: Manual transfer of private key data between devices is required
    - Security concerns: No ability to rotate keys for enhanced protection


These constraints pose significant challenges, especially for newcomers to the crypto space. Users must quickly learn complex security practices to safeguard their assets effectively.

EOA-based systems typically lack robust multi-device support. Users who need to access their accounts on different devices often compromise their security by manually transferring recovery phrases.

Moreover, managing multiple networks, applications, and browser extensions not only adds complexity but also exposes users to security risks such as malware and phishing attacks.

The result is a complex and unforgiving user experience that falls short of the standards set by many modern digital services outside the crypto sphere.



### What are Smart accounts supposed to do

Smart Accounts allow transactions to be approved by multiple authentication methods rather than having to sign every transaction using a private key.

- Support Web 2.0 based standards like FIDO2.

*you can use familiar authentication technologies like Face ID, Touch ID, and Passkeys to easily create a new account, or manage or access existing accounts.*

- Support session based tokens so that we can support 1-click authentication for a short period of time (Sign a payload and create a session for a short time, this will greatly help user experience for trade placement imo)
- You are no longer tied to just a mnemonic.
- Traditional EOAs are still supported.
- Ability to re-register your new FIDO2 device, based on a certain flow (like use your old FIDO2 credentials or  use an EOA.(or maybe even some recovery way via an admin if they have opted in to it, this is more of a nice to have)
