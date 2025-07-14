# Supply-chain rules for Bazel

This repository contains Bazel modules for injecting and collecting supply-chain metadata into builds.

  - [Documentation](./docs)
  - Modules
    - [@package_metadata](./metadata)
  - Contact:
    - [Slack](https://bazelbuild.slack.com/archives/C04AZC3E729)
    - There is a working group which meets weekly on Thursdays at 2:30pm CET / 8:30am EST. [Meet link](https://meet.google.com/qop-eyei-cfh).
      - If you would like to participate, reach out on the slack channel for an invitation.
      - [Meeting notes](https://docs.google.com/document/d/1WhScaOLERet4Fxi4fa2Lpke2MgJZGvEE4EXeq6yb0LU)
    - Mailing list: [bazel-supply-chain-security@bazel.build](https://groups.google.com/a/bazel.build/g/bazel-supply-chain-security)


This project is the successor to [rules_license](https://github.com/bazelbuild/rules_license).

The intended use cases are:
- declaring metadata about packages, such as
  - the licenses the package is available under
  - the canonical package name and version
  - copyright information
  - ... and more TBD in the future
- gathering license declarations into artifacts to ship with code
- applying organization specific compliance constriants against the
  set of packages used by a target.
- producing SBOMs for built artifacts.

WARNING: The code here is still in active initial development and will churn a lot.


## Roadmap

In flux.

### Q3 2025

The immediate concern is feature parity with rules_license and providing a smooth migration path.

## Background reading:


These is for learning about the problem space, and our approach to solutions. Concrete specifications will always appear in checked in code rather than documents.
- [License Checking with Bazel](https://docs.google.com/document/d/1uwBuhAoBNrw8tmFs-NxlssI6VRolidGYdYqagLqHWt8/edit#).
- [OSS Licenses and Bazel Dependency Management](https://docs.google.com/document/d/1oY53dQ0pOPEbEvIvQ3TvHcFKClkimlF9AtN89EPiVJU/edit#)
- [Adding OSS license declarations to Bazel](https://docs.google.com/document/d/1XszGbpMYNHk_FGRxKJ9IXW10KxMPdQpF5wWbZFpA4C8/edit#heading=h.5mcn15i0e1ch)
