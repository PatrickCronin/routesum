# Releasing

* Releasing is done with `goreleaser`. Install it, and check out it's docs.
* Set a `GITHUB_TOKEN` environment variable. See `goreleaser`'s docs for more
  information.
* Update:

  * the copyright dates in `LICENSE.md`, if relevant
  * the copyright dates in `README.md`, if relevant
  * the list of changes, the new version number if one isn't already there, and
    set a release date in `CHANGELOG.md`
  * Commit.

* Tag the release: `git tag -a v1.2.3 -m 'Tag v1.2.3'`.
* Push the tag: `git push origin v1.2.3`.
* Run `goreleaser`.
