# How to Contribute

## Working on several modules

The modules of this repository depend on each other. Their `MODULE.bazel` declare
the minimum versions of the other modules they need once published, and the
`.bazelrc` of each module overrides them with `--override_module`, so that they
are used from this repository, at HEAD, while you work on them.

The paths of the overrides are relative to the directory Bazel is run from, so
run Bazel from the directory of the module (the one with its `MODULE.bazel`),
not from a subdirectory.

The `tests` and `docs` modules are not published, so they keep using
`local_path_override` in their `MODULE.bazel`.

When a module starts to need a newer version of another module, update the
version in its `MODULE.bazel`. Check that the declared versions are enough, with
the overrides ignored:

```sh
% cd tools
% bazel --noworkspace_rc build //...
```

The release workflow does the same before it tags a release.

## Releasing

If you need control over the next release version, for example when making a release candidate for a new major,
then: tag the repo and push the tag, for example

```sh
% git fetch
% git tag v1.0.0-rc0 origin/main
% git push origin v1.0.0-rc0
```

Then watch the automation run on GitHub actions which creates the release.
