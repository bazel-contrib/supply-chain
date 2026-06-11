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
]
