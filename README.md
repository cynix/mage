# Introduction

`mage` is a simple program for encrypting and decrypting multiple files with [age]
using a common passphrase.

[age]: https://age-encryption.org/

# Usage

Simply run `mage` followed by 1 or more files:

    mage path/to/file1.txt etc/file2.conf

    mage path/to/file1.txt.age etc/file2.conf.age

If all files end with `.age`, they will be decrypted. If none of the files end
with `.age`, they will be encrypted. It is not allowed to specify a mixture of
unencrypted and encrypted files. You will be prompted to enter the passphrase.
Input files will be deleted after encryption/decryption.
