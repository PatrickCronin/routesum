# CHANGELOG

## 0.4.0

* Add an iter.Seq output method to rstrie
* Deprecate rstrie.Contents in favor of the iterator method
* Add an iter.Seq output method to routesum
* Deprecate routesum.SummaryStrings in favor of the iterator method
* Prepare rstrie for concurrency

## 0.3.0 (2025-08-17)

* Replace inet.af with net/netip, which removes the `unsafe` dependency
* Output of IPv4-embedded IPv6 networks now uses the convenience form, e.g.
  ::ffff:192.0.2.0 instead of ::ffff:c000:200.

## 0.2.0 (2021-10-19)

* Create a much more performant implementation

## 0.1.1 (2021-09-14)

* Handle empty lines in the CLI tool

## 0.1.0 (2020-12-06)

* Initial beta release
