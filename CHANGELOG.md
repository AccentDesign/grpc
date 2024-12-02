# Change Log
All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.0.26]

* bump go to 1.23.3
* update all dependencies

## [0.0.25]

* bump go to 1.23.2
* update all dependencies

## [0.0.24]

* bump go to 1.23.1
* update all dependencies

## [0.0.23]

* bump go to 1.23
* update all dependencies

## [0.0.22]

* update all dependencies

## [0.0.21]

* bump go to 1.22.5
* update all dependencies

## [0.0.20]

* bump go to 1.22.3
* update all dependencies

## [0.0.19]

* update all dependencies

## [0.0.18]

* bump go to 1.22.2
* update all dependencies

## [0.0.17]

* bump go to 1.22.1
* update all dependencies

## [0.0.16]

* bump go to 1.21.6
* update all dependencies
* remove nested go.mod's

## [0.0.15]

* bump go to 1.21.5
* update all dependencies

## [0.0.14]

* update all dependencies

## [0.0.13]

* update all dependencies

## [0.0.12]

* bump go to 1.21.3
* update all dependencies

## [0.0.11]

* bump go to 1.21.1
* update all dependencies

## [0.0.10]

* bump go to 1.21.0
* update all dependencies

## [0.0.9]

* bump go to 1.20.6
* update all dependencies

## [0.0.8]

* Auth
  * use govalidator.
  * catch gorm.ErrDuplicatedKey in update user
* Email
  * use govalidator.

## [0.0.7]

Add CGO_ENABLED=0 pre-command auth

## [0.0.6]

Add CGO_ENABLED=0 pre-command

## [0.0.5]

* Auth
  * added health service.

* Email
  * added health service.
  * added service validation.

## [0.0.4]

* Email
  * initial test release.

## [0.0.3]

* Auth
  * removed triggers that clear token to internal after update hooks to more easily use diff db's when needed.
  * Added new `-migrations` flag `"on"`, `"off"` and `"dry-run"`.

## [0.0.2]

* Auth
  * added error logging.
  * updated dependencies.

## [0.0.1]

* Auth
    * initial test release.