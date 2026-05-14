"""Custom PURL tests for type-specific validation.

These tests are manually maintained and complement the auto-generated
spec.bzl tests.
"""

visibility([
    "//purl/private/tests/...",
])

custom_tests = [
    # CPAN validation tests
    {
        "description": ("CPAN with lowercase namespace should " +
                        "fail validation"),
        "expected_failure": True,
        "expected_failure_reason": ("CPAN namespace must be " +
                                     "uppercase"),
        "expected_output": None,
        "input": "pkg:cpan/drolsky/DateTime@1.55",
        "test_group": "base",
        "test_type": "parse",
    },
    {
        "description": ("CPAN with mixed case namespace should " +
                        "fail validation"),
        "expected_failure": True,
        "expected_failure_reason": ("CPAN namespace must be " +
                                     "uppercase"),
        "expected_output": None,
        "input": "pkg:cpan/Drolsky/DateTime@1.55",
        "test_group": "base",
        "test_type": "parse",
    },
    {
        "description": ("CPAN build with lowercase namespace " +
                        "should fail validation"),
        "expected_failure": True,
        "expected_failure_reason": ("CPAN namespace must be " +
                                     "uppercase"),
        "expected_output": None,
        "input": {
            "name": "DateTime",
            "namespace": "drolsky",
            "qualifiers": None,
            "subpath": None,
            "type": "cpan",
            "version": "1.55",
        },
        "test_group": "base",
        "test_type": "build",
    },
    # Julia validation tests
    {
        "description": ("Julia with namespace should fail " +
                        "validation"),
        "expected_failure": True,
        "expected_failure_reason": ("Julia PURLs must not have " +
                                     "a namespace"),
        "expected_output": None,
        "input": ("pkg:julia/somenamespace/AWS@1.0.0?" +
                  "uuid=fbe9abb3-538b-5e4e-ba9e-bc94f4f92ebc"),
        "test_group": "base",
        "test_type": "parse",
    },
    {
        "description": ("Julia build with namespace should fail " +
                        "validation"),
        "expected_failure": True,
        "expected_failure_reason": ("Julia PURLs must not have " +
                                     "a namespace"),
        "expected_output": None,
        "input": {
            "name": "AWS",
            "namespace": "somenamespace",
            "qualifiers": {
                "uuid": "fbe9abb3-538b-5e4e-ba9e-bc94f4f92ebc",
            },
            "subpath": None,
            "type": "julia",
            "version": "1.0.0",
        },
        "test_group": "base",
        "test_type": "build",
    },
    {
        "description": ("Julia with both version and uuid " +
                        "should pass"),
        "expected_failure": False,
        "expected_failure_reason": None,
        "expected_output": {
            "name": "AWS",
            "namespace": None,
            "qualifiers": {
                "uuid": "fbe9abb3-538b-5e4e-ba9e-bc94f4f92ebc",
            },
            "subpath": None,
            "type": "julia",
            "version": "1.0.0",
        },
        "input": ("pkg:julia/AWS@1.0.0?" +
                  "uuid=fbe9abb3-538b-5e4e-ba9e-bc94f4f92ebc"),
        "test_group": "base",
        "test_type": "parse",
    },
    {
        "description": ("Julia roundtrip with both version and " +
                        "uuid should pass"),
        "expected_failure": False,
        "expected_failure_reason": None,
        "expected_output": ("pkg:julia/AWS@1.0.0?" +
                            "uuid=fbe9abb3-538b-5e4e-ba9e-bc94f4f92ebc"),
        "input": ("pkg:julia/AWS@1.0.0?" +
                  "uuid=fbe9abb3-538b-5e4e-ba9e-bc94f4f92ebc"),
        "test_group": "base",
        "test_type": "roundtrip",
    },
    # Validation tests for types with normalization rules
    {
        "description": "ALPM build without namespace should fail validation",
        "expected_failure": True,
        "expected_failure_reason": "ALPM PURLs require a namespace",
        "expected_output": None,
        "input": {
            "name": "pacman",
            "namespace": None,
            "qualifiers": None,
            "subpath": None,
            "type": "alpm",
            "version": "6.0.1-1",
        },
        "test_group": "base",
        "test_type": "build",
    },
    {
        "description": "Hackage build with namespace should fail validation",
        "expected_failure": True,
        "expected_failure_reason": "Hackage PURLs must not have a namespace",
        "expected_output": None,
        "input": {
            "name": "AC-HalfInteger",
            "namespace": "haskell",
            "qualifiers": None,
            "subpath": None,
            "type": "hackage",
            "version": "1.2.1",
        },
        "test_group": "base",
        "test_type": "build",
    },
    {
        "description": ("PyPI build with namespace should fail " +
                        "validation using mixed-case type"),
        "expected_failure": True,
        "expected_failure_reason": "PyPI PURLs must not have a namespace",
        "expected_output": None,
        "input": {
            "name": "Django_package",
            "namespace": "python",
            "qualifiers": None,
            "subpath": None,
            "type": "PyPI",
            "version": "1.11.1.dev1",
        },
        "test_group": "base",
        "test_type": "build",
    },
    {
        "description": "Pub build with namespace should fail validation",
        "expected_failure": True,
        "expected_failure_reason": "Pub PURLs must not have a namespace",
        "expected_output": None,
        "input": {
            "name": "characters",
            "namespace": "dart",
            "qualifiers": None,
            "subpath": None,
            "type": "pub",
            "version": "1.2.0",
        },
        "test_group": "base",
        "test_type": "build",
    },
    # Type-specific normalization tests
    {
        "description": ("ALPM build lowercases namespace from " +
                        "normalization rules"),
        "expected_failure": False,
        "expected_failure_reason": None,
        "expected_output": "pkg:alpm/arch/pacman@6.0.1-1?arch=x86_64",
        "input": {
            "name": "pacman",
            "namespace": "Arch",
            "qualifiers": {
                "arch": "x86_64",
            },
            "subpath": None,
            "type": "alpm",
            "version": "6.0.1-1",
        },
        "test_group": "base",
        "test_type": "build",
    },
    {
        "description": "Hackage build applies kebab-case separators",
        "expected_failure": False,
        "expected_failure_reason": None,
        "expected_output": "pkg:hackage/AC-HalfInteger@1.2.1",
        "input": {
            "name": "AC_HalfInteger",
            "namespace": None,
            "qualifiers": None,
            "subpath": None,
            "type": "hackage",
            "version": "1.2.1",
        },
        "test_group": "base",
        "test_type": "build",
    },
    {
        "description": "Pub build normalizes name characters",
        "expected_failure": False,
        "expected_failure_reason": None,
        "expected_output": "pkg:pub/my_package_1@1.0.0",
        "input": {
            "name": "My Package-1",
            "namespace": None,
            "qualifiers": None,
            "subpath": None,
            "type": "pub",
            "version": "1.0.0",
        },
        "test_group": "base",
        "test_type": "build",
    },
    {
        "description": ("PyPI build lowercases name and replaces " +
                        "underscores"),
        "expected_failure": False,
        "expected_failure_reason": None,
        "expected_output": "pkg:pypi/django-package@1.11.1.dev1",
        "input": {
            "name": "Django_package",
            "namespace": None,
            "qualifiers": None,
            "subpath": None,
            "type": "pypi",
            "version": "1.11.1.dev1",
        },
        "test_group": "base",
        "test_type": "build",
    },
    # https://github.com/package-url/purl-spec/pull/793
    {
       "description": "scheme followed by double slash should be accepted",
       "test_group": "base",
       "test_type": "parse",
       "input": "pkg://npm/foo@1.0.0",
       "expected_output": {
           "type": "npm",
           "namespace": None,
           "name": "foo",
           "version": "1.0.0",
           "qualifiers": None,
           "subpath": None
       },
       "expected_failure": False,
       "expected_failure_reason": None
    },
    {
       "description": "scheme followed by triple slash should be accepted",
       "test_group": "base",
       "test_type": "parse",
       "input": "pkg:///pypi/requests@2.28.0",
       "expected_output": {
           "type": "pypi",
           "namespace": None,
           "name": "requests",
           "version": "2.28.0",
           "qualifiers": None,
           "subpath": None
       },
       "expected_failure": False,
       "expected_failure_reason": None
    },
    {
       "description": "scheme with slashes roundtrips to canonical form without slashes",
       "test_group": "base",
       "test_type": "roundtrip",
       "input": "pkg://generic/name@1.0",
       "expected_output": "pkg:generic/name@1.0",
       "expected_failure": False,
       "expected_failure_reason": None
    },
    {
       "description": "plus sign in type should be rejected",
       "test_group": "base",
       "test_type": "parse",
       "input": "pkg:c++/stdlib@1.0",
       "expected_output": None,
       "expected_failure": True,
       "expected_failure_reason": "Should fail to parse a PURL with plus sign in type"
    },
    {
       "description": "plus sign in type should be rejected (build)",
       "test_group": "base",
       "test_type": "build",
       "input": {
           "type": "objective-c++",
           "namespace": None,
           "name": "foundation",
           "version": "1.0",
           "qualifiers": None,
           "subpath": None
       },
       "expected_output": None,
       "expected_failure": True,
       "expected_failure_reason": "Should fail to build a PURL with plus sign in type"
    },
    {
       "description": "colon in name should not be percent-encoded",
       "test_group": "base",
       "test_type": "build",
       "input": {
           "type": "generic",
           "namespace": None,
           "name": "std:io",
           "version": "1.0",
           "qualifiers": None,
           "subpath": None
       },
       "expected_output": "pkg:generic/std:io@1.0",
       "expected_failure": False,
       "expected_failure_reason": None
    },
    {
       "description": "colon in namespace should not be percent-encoded",
       "test_group": "base",
       "test_type": "build",
       "input": {
           "type": "generic",
           "namespace": "org:example",
           "name": "lib",
           "version": None,
           "qualifiers": None,
           "subpath": None
       },
       "expected_output": "pkg:generic/org:example/lib",
       "expected_failure": False,
       "expected_failure_reason": None
    },
    {
       "description": "colon in version should not be percent-encoded",
       "test_group": "base",
       "test_type": "roundtrip",
       "input": "pkg:generic/foo@1.0:beta",
       "expected_output": "pkg:generic/foo@1.0:beta",
       "expected_failure": False,
       "expected_failure_reason": None
    },
    {
       "description": "encoded colon in name should decode and stay unencoded",
       "test_group": "base",
       "test_type": "roundtrip",
       "input": "pkg:generic/std%3Aio@1.0",
       "expected_output": "pkg:generic/std:io@1.0",
       "expected_failure": False,
       "expected_failure_reason": None
    },
    {
       "description": "qualifier with empty value should be discarded",
       "test_group": "base",
       "test_type": "parse",
       "input": "pkg:npm/foo@1.0?empty=&valid=yes",
       "expected_output": {
           "type": "npm",
           "namespace": None,
           "name": "foo",
           "version": "1.0",
           "qualifiers": {
           "valid": "yes"
         },
           "subpath": None
       },
       "expected_failure": False,
       "expected_failure_reason": None
    },
    {
       "description": "multiple empty qualifiers should all be discarded",
       "test_group": "base",
       "test_type": "roundtrip",
       "input": "pkg:npm/bar@2.0?a=&b=value&c=",
       "expected_output": "pkg:npm/bar@2.0?b=value",
       "expected_failure": False,
       "expected_failure_reason": None
    },
    {
       "description": "all-empty qualifiers should result in no query string",
       "test_group": "base",
       "test_type": "roundtrip",
       "input": "pkg:npm/baz@3.0?empty=&also_empty=",
       "expected_output": "pkg:npm/baz@3.0",
       "expected_failure": False,
       "expected_failure_reason": None
    },
    {
       "description": "namespace segments should preserve literal slashes",
       "test_group": "base",
       "test_type": "build",
       "input": {
           "type": "maven",
           "namespace": "org.apache/commons",
           "name": "lang",
           "version": "3.12",
           "qualifiers": None,
           "subpath": None
       },
       "expected_output": "pkg:maven/org.apache/commons/lang@3.12",
       "expected_failure": False,
       "expected_failure_reason": None
    },
    {
        "description": "special characters in namespace segments should be encoded per-segment",
        "test_group": "base",
        "test_type": "build",
        "input": {
           "type": "generic",
           "namespace": "org name/sub dir",
           "name": "lib",
           "version": None,
           "qualifiers": None,
           "subpath": None
        },
        "expected_output": "pkg:generic/org%20name/sub%20dir/lib",
        "expected_failure": False,
        "expected_failure_reason": None
    },
    {
        "description": "namespace with encoded slash in segment should roundtrip correctly",
        "test_group": "base",
        "test_type": "roundtrip",
        "input": "pkg:generic/a%2Fb/c/name",
        "expected_output": "pkg:generic/a%2Fb/c/name",
        "expected_failure": False,
        "expected_failure_reason": None
    }
]
