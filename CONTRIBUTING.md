# How to Contribute

## Releasing

If you need control over the next release version, for example when making a release candidate for a new major,
then: tag the repo and push the tag, for example

```sh
% git fetch
% git tag v1.0.0-rc0 origin/main
% git push origin v1.0.0-rc0
```

Then watch the automation run on GitHub actions which creates the release.
