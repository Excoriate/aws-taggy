# Changelog

## [1.4.0](https://github.com/Excoriate/aws-taggy/compare/v1.3.0...v1.4.0) (2025-02-04)


### Features

* Update Homebrew tap token in .goreleaser.yml ([#12](https://github.com/Excoriate/aws-taggy/issues/12)) ([78bd805](https://github.com/Excoriate/aws-taggy/commit/78bd805ebc2ca60b8ec98a3df1a99c3bef74c1e3))

## [1.3.0](https://github.com/Excoriate/aws-taggy/compare/v1.2.1...v1.3.0) (2025-02-04)


### Features

* Add support for AWS SQS resources ([#10](https://github.com/Excoriate/aws-taggy/issues/10)) ([9ac3047](https://github.com/Excoriate/aws-taggy/commit/9ac304742d241bcda2e89c0f36b5d74dbd8d3d5b))

## [1.2.1](https://github.com/Excoriate/aws-taggy/compare/v1.2.0...v1.2.1) (2025-02-04)


### Bug Fixes

* Update Homebrew token in .goreleaser.yml ([#8](https://github.com/Excoriate/aws-taggy/issues/8)) ([aaae42e](https://github.com/Excoriate/aws-taggy/commit/aaae42e5f6959daa30f5e12f716990632417034e))

## [1.2.0](https://github.com/Excoriate/aws-taggy/compare/v1.1.0...v1.2.0) (2025-02-03)


### Features

* Add build-release job to create release artifacts ([#6](https://github.com/Excoriate/aws-taggy/issues/6)) ([8d47b63](https://github.com/Excoriate/aws-taggy/commit/8d47b6310b625e2b8586c4a39b9a7e4d960f5a39))

## [1.1.0](https://github.com/Excoriate/aws-taggy/compare/v1.0.0...v1.1.0) (2025-02-03)


### Features

* Simplify VPCInspector creation by removing redundant comment ([#3](https://github.com/Excoriate/aws-taggy/issues/3)) ([1720ffc](https://github.com/Excoriate/aws-taggy/commit/1720ffc1831a41f7f76c8d55e87b2cc965f85f02))

## 1.0.0 (2025-02-03)


### Features

* add complete validation on missing checks ([3ad7db2](https://github.com/Excoriate/aws-taggy/commit/3ad7db25f6a748fa5d43a2534511af4a4593b0d3))
* add configuration's extra logic ([0105da1](https://github.com/Excoriate/aws-taggy/commit/0105da168e0c37a89fd652e02efdcdcbc12f30f3))
* add discover command ([6d124fc](https://github.com/Excoriate/aws-taggy/commit/6d124fce9a519eba1e3466c063f941988f5a35b5))
* add examples ([8c28e30](https://github.com/Excoriate/aws-taggy/commit/8c28e30d706eab39ef0a544c1f85ea5877a15fb3))
* add fetch command ([0fd47ea](https://github.com/Excoriate/aws-taggy/commit/0fd47ea47c1ca46328e36fc63d457c60e391cc8c))
* add inspector for route53 ([6ceef97](https://github.com/Excoriate/aws-taggy/commit/6ceef97bcc16f5f14e2ccc7cd163e84dbca15656))
* add new cloudwatchlogs inspector ([961f4d3](https://github.com/Excoriate/aws-taggy/commit/961f4d3df332d3862efe3e94951ad558769ec1a5))
* add new cost, and usage functionality in the inspector package ([58ad1c4](https://github.com/Excoriate/aws-taggy/commit/58ad1c449a9f837d753a523d2424041dd7295305))
* add new inspectors ([ec42ba2](https://github.com/Excoriate/aws-taggy/commit/ec42ba2c475a96dcdc77c26834f6d3f13754d2de))
* add output support ([d0c93ff](https://github.com/Excoriate/aws-taggy/commit/d0c93ff6cabc91861629e4e518fb3a15a4a7e322))
* add project initial structure ([7aee6cf](https://github.com/Excoriate/aws-taggy/commit/7aee6cf59bb72223088c7beeda86195cefcceb8e))
* add scanner ([9ffe663](https://github.com/Excoriate/aws-taggy/commit/9ffe66388e2ead087c366dcccdad59b43b25cb00))
* add second scenario, for non compliant resource ([70b70bc](https://github.com/Excoriate/aws-taggy/commit/70b70bcbd6b6a661a0b7ce2618b569571dc81a1e))


### Bug Fixes

* add multiple violations per run ([d555480](https://github.com/Excoriate/aws-taggy/commit/d5554801b73b0c88eae69aaa053036e9618611bf))
* added missing validations ([8fbf7d6](https://github.com/Excoriate/aws-taggy/commit/8fbf7d6e2b52fab981a786584aea4da90cc8d086))
* adjust query command ([dcd3e10](https://github.com/Excoriate/aws-taggy/commit/dcd3e10f40c76e7d3428919da95b6e05b4327baa))
* amend tests ([ab55a4b](https://github.com/Excoriate/aws-taggy/commit/ab55a4b12d2865340568a7e1e9e051ceb649b002))
* deadlock on goroutine ([2f0be1b](https://github.com/Excoriate/aws-taggy/commit/2f0be1ba37021ba660fadceef415bf1fe9e484ea))
* deadlock on goroutine ([a2bfc47](https://github.com/Excoriate/aws-taggy/commit/a2bfc47c3d2fbbeb7fd9a6b23571526475861641))
* failed tests on windows ([61bce00](https://github.com/Excoriate/aws-taggy/commit/61bce00b15345735800363b3a6120042e03a9cc6))
* fix tests ([51b5f76](https://github.com/Excoriate/aws-taggy/commit/51b5f76c4df7e411edd7db67520f3f2c31fddf6b))
* golangci on ci ([de3e999](https://github.com/Excoriate/aws-taggy/commit/de3e99942319f477354255b2e9cb6caf410a0b1e))
* gowork error, add configuration package ([a1ed54c](https://github.com/Excoriate/aws-taggy/commit/a1ed54c54c3e370c83014bf8d44f1aa936a48bd3))
* new compliance check command ([6322087](https://github.com/Excoriate/aws-taggy/commit/63220875912b527badedae241f78357e6e71edc4))
* scenario 1 ([a7652b5](https://github.com/Excoriate/aws-taggy/commit/a7652b5ce4e528f0ed4132cc5aa730995936f83f))
* tests ([7736873](https://github.com/Excoriate/aws-taggy/commit/7736873ad21a7490539829bd70f3d668d4ef6aa7))
