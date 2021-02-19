# Security Policy

## Reporting Security Issues

The Provenance team and community take all security vulnerabilities in Provenance seriously. 

Thank you for improving the security of Provenance. We appreciate your efforts and responsible disclosure and will make every effort to acknowledge your contributions.

Please report security vulnerabilities to
**[security@provenance.io](mailto:security@provenance.io)**.  *Please avoid opening a public issue on the repository for security issues.*


The Provenance team will send a response indicating the next steps in handling your
report. After the initial reply to your report, the team will keep you informed
of the progress towards remediation and may ask for additional
information or guidance.  For critical problems you may encrypt your report using the public key below.

In addition, please include the following information along with your report:

* Your name and affiliation (if any).
* A description of the technical details of the vulnerabilities. It is very important to let us know how we can reproduce your findings.
* An explanation of who can exploit this vulnerability, and what they gain when doing so -- write an attack scenario. This will help us evaluate your report quickly especially if the issue is complex.
* Whether this vulnerability is public or known to third parties. If it is, please provide details.

If you believe that an existing (public) issue is security-related, please send
an email to **[security@provenance.io](mailto:security@provenance.io)**. The email should include the issue ID and
a short description of why it should be handled according to this security
policy.

If you have found an issue with the Cosmos SDK or Tendermint modules not found in this repo you can report them through links found here. https://tendermint.com/security/

## Disclosure Policy

When the security team receives a security bug report, they will assign it to a primary handler. This person will coordinate the fix and release process, involving the following steps:

* Confirm the problem and determine the affected versions.
* Audit code to find any potential similar problems.
* Prepare fixes for all releases still under maintenance. These fixes will be released as fast as possible to the project.


#### Encryption key for `security@provenance.io`

If your disclosure is sensitive, you may choose to encrypt your
report using the key below. 
Please only use this for critical security
reports.

```
-----BEGIN PGP PUBLIC KEY BLOCK-----

mQINBGAtlcIBEADAb3umrdeMNdG i8pbv49K0R6XKRab7Gl2cAYS7SvLWJgBBh94JsmQIXID4vORI/XOcq/rwAGQ/Z7fGx/pPesmYRTkSY7G0 JKZRIkfXBuTS5QufAQDNBVd79u8lkbcbhtnrj56uaVOmcSiMeSLBHNBAo/RjtyV/XZ31w5tgwEczRJslAJf9exwgH9/Mpvl3lihEAmmAt838a5SzTZCeiYJg/tdmMMe5wEjl6dSkJ/LhfITectQAnlniiKZnOvNBsnw0mICZU18sKvvTQPsHRJl9bITcG0maSsbpuW0gb3ysu54Lp02fMnS5PzpI DefdGi3oepre LnolrSYLxaB0urzjAgsPM9xrktISoxRnhBCl8OFQC//Nyr0ru47MBWm/VYAFRRQrol0QrFg6U07/CnaSrTBlq8D9Q9kJiC5diJmkgg1y2 yzp iHX08CUJsDhqEBv4jKdAJYh t8xja32FqHsiOdiqbSa89sO7KjiJu1UkmF8e3X wDIIGpfKGgbOFvZuk7GJDIPNNeGudt8TOn4Bphdn3xTfrlvmc3vIzCzn7Doh0Mdgf/nDTNEBffbi9xKePN8bpYmmt7C1FYYiAjYodNSRp2SaBbzdHchbFu2IAeKrdJ5QgEdRt3DdnUD3mOm/wpzYB2xZLUngqhJt1AISBu8hMAWbEYKyXKSTwARAQABtCVGaWd1cmUgU2VjdXJpdHkgPHNlY3VyaXR5QGZpZ3VyZS5jb20 iQJUBBMBCAAFiEEuSeVJQji3fEZHijh2kfGatENIL0FAmAtlcICGwMFCQeGHcoFCwkIBwIGFQoJCAsCBBYCAwECHgECF4AACgkQ2kfGatENIL2 QhAAvFFl14tiiI LJOZFm6G4ig j3g5iA3R1V7aQs7PvMK4NpmARk3Ctd 7LrEJBVrYgsuBhwQTpG5i4cyl0JaYd DH3r g2/pqOGxMbgfir1ZLKsN0IZnkeeyclHAQVOHY9HiaCvSCa/aEh6smKqJ7kY8zbjssxOK66Llldfl9b31FcFz5uv1i6yLpfTG8E1 JhslJPzaUgbyildy67GSlNx5pzP0ib7Lk 75ptgGUMTyLm50yV6DoTIQNVd3R4mHbnZTp 0qq04/xR9KWzoVP6VUmHx63 ER6LQu53Gl8BEMQ4TjWEV DiDeT2Th6ehS2jeCd5uID QpZxNLjfXv45t6CxNYBJDNXG25kfPGYMKN8pxleBfYAidsTjGasLfghSQDzenrH7JvUpdDNsnP5Ftws8FUyfzhOVBdXjvmrNYeRgtE5baoeqgn2p8gh5Ug4dJmuhmZavmHmQi3 WXx146fNQxN2crs371nUdbn1ASiDfunOT/5VPG7coShAeq1sJ20wj9drK8jg9aFRh0B9hGYwqUUF1Lu72uYtREy6Jjo/qX9DM/C/uUjtWQj5nT/58ljqW 3Ezc3ZrPC1s7l5diJh0iZWGL1jxLg1Ks6AeiACFipaKkgbyT62b9yiRtjspH01ZjfACojMaFMoVdcPBCB JgaszFcXyRARl0BFWxq 5Ag0EYC2VwgEQANPmOgiD8LsRRf AzNP3qwIj8SPs62Y2B4JCrz9UMAimIw3tnXriVubIdXz19/Y8hGUP4xqyI6IQnJG7k66U28XqOK6spsygFuerwyCCkyzCr5QQJPYQjocqv4JDkEyjHKNofHz0ZWS7xnI 5zI3MDFErszQIIxwiF1t7lmN0mHzA/3C1O0JyXmtwCRs9VikoABFXp5Xbpg/dsmlZNyBp6Mti1MHjCWQTynoe3wzsP OZNLY2uQQW5Sd1Bh7rhQty8RZbIeyovkOh3Gj7SDZwgwKVEYse8Ef7CfIGVWL ULoDdcEopA8tL/xM31f6E5YwJEorUcHa85hWCJ0 jhl053eAm2PZ1AJEfgVwe1sl O9uKsp5Z8MG L6A278mwYcrskFTP6aJWu1DNr84tJgDcedIIWcVRWhrYrBv3vcm1DpK39Rv0SOLcPS9tj3WeQ/DW4/rvkCm8H8NTXU5HQzCTWH42/RqFA/R9s2Uieq8/oUAtbis4luLE9m6osTXCHZoQehmKSqnU9ho1oR70p6Mf33SIRUDAJvAcALAG9wJ0cxFo1r AoV8YfGLAbaVFI5pGlNouZZhJUuUkH9C0tIdL3GzOhwybYurKhbKoZBw44fLtvKxeP55q9gWkJAG/vDrjdVmhrEnDDqv9bqQatPuC3PO8w/MnNTwxVGghPZUYAPABEBAAGJAjwEGAEIACYWIQS5J5UlCOLd8RkeKOHaR8Zq0Q0gvQUCYC2VwgIbDAUJB4YdygAKCRDaR8Zq0Q0gvX2HD/4nLFGIvzqnoUB/ FSRrwFJ0b53lNLWmbUpVB1sm1oiJkJtFRbUAAbeyoyUjL6xcE5x640mg39J8tjyvPdR1s7hO0bSvBxSM/IEn329YQvj/eDV8pEXQBUWTeCcCHp8BxuJROBdIeu3fjO4fyywIm45QnFTCnrPNS2mA9brg8w D6Zo1LOhBto5OvOgirV2K9v6koI0/070IuCt3 oKsVaEMer8f/Ta3h05r810PRzqgWvGmQgyjkXln7rAg8cRytC/i3wEdSTiV 7NiBsY/jQcPu5DEncFx8bA1ZzYAPGQ90g6FLHhMFUo/uT9YHYH2 vglTVTSDelJqt3SlVNOL/ Czfxg24eNCdthe2yjsLskkQw1pstAStLu7eTUu1swdkrq9T3r6J7l9vvYwdD6R6r5bEpr ghG567RIIbofzciLIkMvxLROussb4VjhhahT/l0nfFYvoYQw7OPRA7z1RzHV3iaIfWMqvnf6E/fL3F0AKj7ff8mRd9ueX0LySAe5f28qe6vDEio2kKkDWWygTgphL5dGsb2KOuOwyqvzOYiKcAwHF4Wl3Gju5SH3S5 2O4Tvs2HlEPjJ6JLb1wxYVy1WG44DUthmzd0/M553j1ZcIRMfBDtVlfXVliGvwNeLeEO/lHLJxPr F85RFTCgnj/diuIjESeNXtni6gAgzS5g===0lzZ
-----END PGP PUBLIC KEY BLOCK-----
```

## Comments on this Policy

If you have suggestions on how this process could be improved please submit a pull request or email **[security@provenance.io](mailto:security@provenance.io)**
