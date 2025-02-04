# Changelog - Paw

## 0.15.0 - 26 January 2022 

- cli: add CLI application #3

## 0.14.0 - 09 January 2022

> This release updates the internal storage, so previous versions won't be compatible.
> Starting from this release the data is encoded in json in place of gob 
> and update to use a password protected age key (X25519) to decrypt and encrypt the vault data.
> This allow to decrypt the items using the age tool and have the content directly accessible once decrypted.

- paw: update to use a password protected age key (X25519) to decrypt and encrypt the vault data
- paw: data encoded in json in place of gob
- paw,ui: group vault ItemMetadata by ItemType
- paw,ui: export item UX improvement: items are now decoded concurrently and a progress bar is shown if needed
- ui: show item count into the item select list
- ui: fix renaming an item when a filter is specified could display the vault empty view

## 0.13.1 - 07 January 2022

- paw: item creation was not working correctly
- doc: update screenshot

## 0.13.0 - 07 January 2022

> This release updates the internal storage, so previous versions won't be compatible.

- paw: the website item has be renamed into login to make it more general purpose
- ui: support showing website favicons #8
- favicon: add package favicon that provides a favicon downloader #8
- ico: add package ico that implements a minimal ICO image decoder #8

## 0.12.0 - 03 January 2022

> This release updates the internal storage, so previous versions won't be compatible.

- paw,ui: import items from file #6
- paw,ui: export items to file #7
- haveibeenpwned,ui: improve password audit

## 0.11.0 - 28 December 2021

> This release updates the internal storage, so previous versions won't be compatible.

- paw: items are now stored into dedicated age files
- paw,ui: add passphrase support #4
- paw,ui: add pin password support

## 0.10.0 - 21 December 2021

### Added

- Password audit against data breaches #1 
- Add TOTP and HTOP support #5

## 0.9.0 - 11 December 2021

- First release
