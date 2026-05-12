# Supply-chain rules for Bazel

This repository contains Bazel modules for injecting and collecting supply-chain metadata into builds.

  - [Documentation](./docs)
  - Modules
    - [@package_metadata](./metadata)

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

> [!WARNING]
> The code here is still in active initial development and will churn a lot.

## How to participate

- [Slack](https://bazelbuild.slack.com/archives/C04AZC3E729)
- Mailing list: [bazel-supply-chain-security@bazel.build](https://groups.google.com/a/bazel.build/g/bazel-supply-chain-security)
- Weekly meeting every Thursday at 8:30am EST / 02:30pm CET ([Add to calendar](https://calendar.google.com/calendar/event?action=TEMPLATE&tmeid=MXA1ZXJlZ3Mxa24xcTl1c3NocXQ1dnFwNXRfMjAyNTEyMThUMTMzMDAwWiBjXzYxNTYxMjA0MmE4YjUyODgxYWJkYjk3NDcyNDdmZDEyYjA5NDhhYWFhNTFkNDE5YmUzNWE3ODgyNWRkM2RhNmRAZw&tmsrc=c_615612042a8b52881abdb9747247fd12b0948aaaa51d419be35a78825dd3da6d%40group.calendar.google.com&scp=ALL) / [meet ID](https://meet.google.com/qop-eyei-cfh) / [meeting notes](https://docs.google.com/document/d/1WhScaOLERet4Fxi4fa2Lpke2MgJZGvEE4EXeq6yb0LU/edit?usp=sharing))

## Roadmap

See [this page](roadmap.md).

## Background reading:

These is for learning about the problem space, and our approach to solutions. Concrete specifications will always appear in checked in code rather than documents.
- [License Checking with Bazel](https://docs.google.com/document/d/1uwBuhAoBNrw8tmFs-NxlssI6VRolidGYdYqagLqHWt8/edit#).
- [OSS Licenses and Bazel Dependency Management](https://docs.google.com/document/d/1oY53dQ0pOPEbEvIvQ3TvHcFKClkimlF9AtN89EPiVJU/edit#)
- [Adding OSS license declarations to Bazel](https://docs.google.com/document/d/1XszGbpMYNHk_FGRxKJ9IXW10KxMPdQpF5wWbZFpA4C8/edit#heading=h.5mcn15i0e1ch)
