# Roadmap

This page covers our roadmap for 2026.

## Milestone M1.0 - BazelCon minus 3 months (July 2026)

- Supply chain core
    - Get package_metadata to 1.0 (mostly done)
        - Finish purl work
        - Need to update documentation to remove experimental disclaimer
    - Get tooling module to 1.0
        - Antonio is working on proposal to redo tooling
        - Should we split it into a third module?
            - Package_metadata, supply-chain-go, supply-chain-tools (+ language-specific libraries)
    - Bzlmod integration
- Adoption
    - Get 1.0 out to actual users
    - Every ruleset under bazel-contrib that pulls packages should use package_metadata
    - Add package_metadata to some core rule sets (cc, java, python, shell, skylib, bazel-features, aspect rules, rules_pkg)
    - Have something in the BCR to check (later enforce) presence of package_metadata
    - Identify early adopters
        - EngFlow customers?
        - Google-owned OSS projects? Envoy, Pigweed

## Milestone M1.1 - BazelCon (October 2026)

- Incorporate dogfood user feedback
- Add missing attributes that are ecosystem-specific (or have a good story at least)
    - Should they live in package_metadata? Centralized storage, or rather in the rule set
- Enforce attributes
- Bazel itself should use package_metadata e2e to produce a complete SBOM, should be part of all builds
    - Need to track all bzl files involved in a build, including attestations from module metadata and directly from Bazel. Which module extensions were involved?
    - "No dependency should go untracked for a build, unless explicitly opted-out" -> should eventually fail the build
    - We need accountability on the action level, not just the target graph level

## Future Milestones (post-BazelCon)

- Integration for signing SBOMs
- Providing secure provenance
- How to integrate this with other security measures?
- Action cache signing is related (https://github.com/bazelbuild/remote-apis/issues/368) - maybe Ulf wants to work on this?
- Attest to every byte of an artifact and its origin
