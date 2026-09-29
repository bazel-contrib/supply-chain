# Bazel Central Registry

This directory contains configuration files to automatically publish releases
(`"git tags"`) to the [Bazel Central Registry](https://registry.bazel.build).

Things to note:
- There is 1 folder per sub-release from the project
- The folder is named by the **published** release (and BCR) name
- The presubmit job uses the published name because it has to match
  what an end user would see.
- The source template inside each uses the source tree **folder** name
  because it must provide the correct strip_prefix based on the tarball.

See <https://github.com/bazel-contrib/publish-to-bcr/blob/main/templates/README.md>
for authoritative documentation about these files.
