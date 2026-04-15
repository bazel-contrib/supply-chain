"""Module defining a builder for [purl](https://github.com/package-url/purl-spec)s."""

load("//purl/private/normalization:normalization.bzl", "normalize")
load("//purl/private/percent_encoding:percent_encoding.bzl", "percent_encode")
load("//purl/private/validation:validation.bzl", "validate")

visibility([
    "//purl/...",
])

def _type(self, fields, type_name):
    fields["type"] = type_name
    return self

def _namespace(self, fields, namespace):
    fields["namespace"] = namespace
    return self

def _name(self, fields, name):
    fields["name"] = name
    return self

def _version(self, fields, version):
    fields["version"] = version
    return self

def _add_qualifier(self, fields, name, value):
    fields.setdefault("qualifiers", {})[name] = value
    return self

def _subpath(self, fields, subpath):
    fields["subpath"] = subpath
    return self

def _build(self, fields):
    purl, err = build(
        type = fields.get("type", None),
        namespace = fields.get("namespace", None),
        name = fields.get("name", None),
        version = fields.get("version", None),
        qualifiers = fields.get("qualifiers", None),
        subpath = fields.get("subpath", None),
    )

    if err:
        fail(err)
    return purl

def _is_type(actual, expected):
    return type(actual) == type(expected)

def build(
        *,
        type = None,
        namespace = None,
        name = None,
        version = None,
        qualifiers = {},
        subpath = None):
    """Builds a Package URL (PURL) string from component parts.

    This function validates, normalizes, and serializes the PURL components
    according to the PURL specification (https://github.com/package-url/purl-spec).

    Args:
        type: The package type (required). Must be lowercase ASCII string (e.g., "maven", "npm", "pypi").
        namespace: The package namespace (optional). Can be a string or list of strings for multi-segment namespaces.
        name: The package name (required). Will be percent-encoded in the output.
        version: The package version (optional). Will be percent-encoded in the output.
        qualifiers: A dictionary of qualifier key-value pairs (optional). Keys must start with ASCII letter
                    and contain only lowercase letters, numbers, '.', '-', '_'. Values will be percent-encoded.
        subpath: A list of path segments (optional). Each segment will be percent-encoded.

    Returns:
        A tuple of (purl_string, error). On success, returns (purl_string, None).
        On failure, returns (None, error_message).

    Example:
        ```starlark
        purl, err = build(
            type = "maven",
            namespace = "org.apache.commons",
            name = "commons-lang3",
            version = "3.12.0",
        )
        # purl: pkg:maven/org.apache.commons/commons-lang3@3.12.0
        # err: None
        ```
    """

    err = validate(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
    if err:
        return None, err

    purl, err = normalize(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
    if err:
        return None, err

    # Serialization accoring to https://github.com/package-url/purl-spec/blob/aaede64286deb66c19a80974397d2d903c393d64/docs/how-to-build.md and Section 5.6 of https://ecma-international.org/wp-content/uploads/ECMA-427_1st_edition_december_2025.pdf.
    components = []

    # Start a PURL string with the scheme as a lowercase ASCII string.
    components.append("pkg:")

    # Append the type string to the PURL as an unencoded lowercase ASCII string.
    #   - Append '/' to the PURL.
    components.append(purl.type)
    components.append("/")

    # If the namespace is not empty:
    if purl.namespace:
        # Percent-encode each segment.
        segments = [percent_encode(v) for v in purl.namespace]

        # Join the segments with '/'.
        # Append this to the PURL.
        components.append("/".join(segments))

        # Append '/' to the PURL.
        components.append("/")

    # Append the percent-encoded name to the PURL.
    components.append(percent_encode(purl.name))

    # If the version is not empty:
    if purl.version:
        # Append '@' to the PURL.
        components.append("@")

        # Append the percent-encoded version to the PURL.
        components.append(percent_encode(purl.version))

    # If the qualifiers are not empty and not composed only of key/value pairs
    # where the value is empty:
    if purl.qualifiers:
        # Append '?' to the PURL.
        components.append("?")

        # Build a list from all key/value pair:
        pairs = []
        for key, v in purl.qualifiers.items():
            # If the key is 'checksum' and this is a list of checksums join this
            # list with a ',' to create this qualifier value.
            if (key == "checksum") and _is_type(v, []):
                value = ",".join(v)
            else:
                value = v

            # Create a string by joining the lowercased key, the equal '=' sign
            # and the percent-encoded value to create a qualifier.
            pairs.append("{}={}".format(key, percent_encode(value)))

        # - Sort this list of qualifier strings lexicographically.
        # - Join this list of qualifier strings with a '&' ampersand.
        # - Append this string to the PURL.
        components.append("&".join(sorted(pairs)))

    # If the subpath is not empty and not composed only of empty, '.' and '..' segments:
    if purl.subpath:
        # Append '#' to the PURL.
        components.append("#")

        # Percent-encode each segment.
        segments = [percent_encode(segment) for segment in purl.subpath]

        # - Join the segments with '/'.
        # - Append this to the PURL.
        components.append("/".join(segments))

    return "".join(components), None

def builder():
    """Creates a fluent builder for constructing Package URLs (PURLs).

    The builder uses a fluent API pattern where methods can be chained together.
    Call `build()` at the end to construct the final PURL string.

    Example:
        ```starlark
        purl = builder() \
            .type("maven") \
            .namespace("org.apache.commons") \
            .name("commons-lang3") \
            .version("3.12.0") \
            .build()
        # Result: pkg:maven/org.apache.commons/commons-lang3@3.12.0
        ```

    Example with qualifiers:
        ```starlark
        purl = builder() \
            .type("npm") \
            .name("express") \
            .version("4.18.2") \
            .add_qualifier("arch", "x64") \
            .add_qualifier("os", "linux") \
            .build()
        # Result: pkg:npm/express@4.18.2?arch=x64&os=linux
        ```

    Example with subpath:
        ```starlark
        purl = builder() \
            .type("github") \
            .namespace("curl") \
            .name("curl") \
            .version("7.72.0") \
            .subpath(["lib", "vtls"]) \
            .build()
        # Result: pkg:github/curl/curl@7.72.0#lib/vtls
        ```

    Returns:
        A builder object with the following methods:
            - `type(type_name)`: Sets the package type (required). Must be lowercase ASCII.
            - `namespace(namespace)`: Sets the namespace. Can be a string or list of strings for multi-segment namespaces.
            - `name(name)`: Sets the package name (required).
            - `version(version)`: Sets the package version (optional).
            - `add_qualifier(name, value)`: Adds a qualifier key-value pair (optional). Key must start with ASCII letter and contain only lowercase letters, numbers, '.', '-', '_'.
            - `subpath(subpath)`: Sets the subpath. Must be a list of strings representing path segments (optional).
            - `build()`: Constructs and returns the PURL string. Fails if validation errors occur.
    """

    fields = {}
    self = struct(
        type = lambda type_name: _type(self, fields, type_name),
        namespace = lambda namespace: _namespace(self, fields, namespace),
        name = lambda name: _name(self, fields, name),
        version = lambda version: _version(self, fields, version),
        add_qualifier = lambda name, value: _add_qualifier(self, fields, name, value),
        subpath = lambda subpath: _subpath(self, fields, subpath),
        build = lambda: _build(self, fields),
    )
    return self
