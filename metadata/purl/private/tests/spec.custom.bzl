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
]
