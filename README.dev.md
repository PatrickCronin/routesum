# Releasing

* Releasing is done with `goreleaser`. Install it, and check out it's docs.
* Set a `GITHUB_TOKEN` environment variable. See `goreleaser`'s docs for more
  information.
* Update `CHANGELOG.md`:

  * Mention recent changes.
  * Set a version if there is not one.
  * Set a release date.
  * Commit

* Tag the release: `git tag -a v1.2.3. -m 'Tag v1.2.3'`.
* Push the tag: `git push origin v1.2.3`.
* Run `goreleaser`.
