"""Module defining a builder for [purl](https://github.com/package-url/purl-spec)s."""

load("//purl/private/normalization:normalization.bzl", "normalize")
load("//purl/private/percent_encoding:percent_encoding.bzl", "percent_encode")
load("//purl/private/validation:validation.bzl", "validate")

visibility([
    "//purl/...",
])

def _type(self, fields, type_name):
    """Sets the package type (required).

    Args:
        type_name: The package type as a lowercase ASCII string (e.g., "maven", "npm", "pypi").

    Returns:
        The builder instance for method chaining.
    """
    fields["type"] = type_name
    return self

def _namespace(self, fields, namespace):
    """Sets the package namespace (optional).

    Args:
        namespace: The namespace as a string with segments separated by '/' (e.g., "org.apache.commons").
                   Multi-segment namespaces like "github.com/user/project" are supported.

    Returns:
        The builder instance for method chaining.
    """
    fields["namespace"] = namespace
    return self

def _name(self, fields, name):
    """Sets the package name (required).

    Args:
        name: The package name. Will be percent-encoded in the final PURL.

    Returns:
        The builder instance for method chaining.
    """
    fields["name"] = name
    return self

def _version(self, fields, version):
    """Sets the package version (optional).

    Args:
        version: The package version. Will be percent-encoded in the final PURL.

    Returns:
        The builder instance for method chaining.
    """
    fields["version"] = version
    return self

def _add_qualifier(self, fields, name, value):
    """Adds a qualifier key-value pair (optional, repeatable).

    Args:
        name: The qualifier key. Must start with an ASCII letter and contain only
              lowercase letters, numbers, '.', '-', '_'.
        value: The qualifier value. Will be percent-encoded in the final PURL.

    Returns:
        The builder instance for method chaining.
    """
    fields.setdefault("qualifiers", {})[name] = value
    return self

def _subpath(self, fields, subpath):
    """Sets the subpath (optional).

    Args:
        subpath: The subpath as a string with segments separated by '/' (e.g., "src/main").
                 Each segment will be percent-encoded in the final PURL.

    Returns:
        The builder instance for method chaining.
    """
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

    Validation and normalization are performed in two stages:
    1. General validation checks required fields and qualifier key constraints
    2. Type-specific validation and normalization rules are applied based on the
       package type (e.g., case normalization for npm, path handling for golang)

    For a list of supported PURL types and their specifications, see:
    https://github.com/package-url/purl-spec/blob/main/purl-types-index.json

    Args:
        type: The package type (required). Must be lowercase ASCII string (e.g., "maven", "npm", "pypi").
        namespace: The package namespace (optional). String with segments separated by '/' (e.g., "org.apache.commons").
        name: The package name (required). Will be percent-encoded in the output.
        version: The package version (optional). Will be percent-encoded in the output.
        qualifiers: A dictionary of qualifier key-value pairs (optional). Keys must start with ASCII letter
                    and contain only lowercase letters, numbers, '.', '-', '_'. Values will be percent-encoded.
        subpath: The subpath (optional). String with segments separated by '/' (e.g., "src/main").

    Returns:
        A tuple of (purl_string, error). On success, returns (purl_string, None).
        On failure, returns (None, error_message).

    Example:
        ```starlark
        purl, err = build(
            type = "maven",
            namespace = "org.apache.xmlgraphics",
            name = "batik-anim",
            version = "1.9.1",
            qualifiers = {
                "classifier": "sources",
                "repository_url": "https://repo.spring.io/release",
            },
        )
        # purl: pkg:maven/org.apache.xmlgraphics/batik-anim@1.9.1?classifier=sources&repository_url=https%3A%2F%2Frepo.spring.io%2Frelease
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
        # Sort by keys first, then build the pairs
        pairs = []
        for key in sorted(purl.qualifiers.keys()):
            v = purl.qualifiers[key]

            # If the key is 'checksum' and this is a list of checksums join this
            # list with a ',' to create this qualifier value.
            if (key == "checksum") and _is_type(v, []):
                value = ",".join(v)
            else:
                value = v

            # Create a string by joining the lowercased key, the equal '=' sign
            # and the percent-encoded value to create a qualifier.
            pairs.append("{}={}".format(key, percent_encode(value)))

        # - Join this list of qualifier strings with a '&' ampersand.
        # - Append this string to the PURL.
        components.append("&".join(pairs))

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

    The builder provides a chainable interface for constructing PURLs according to
    the [Package URL specification](https://github.com/package-url/purl-spec).

    The `type` and `name` fields are required. All components are validated and
    normalized according to the PURL spec. Components are automatically percent-encoded
    where necessary, and qualifiers are sorted lexicographically in the output.

    For a list of supported PURL types and their specifications, see:
    https://github.com/package-url/purl-spec/blob/main/purl-types-index.json

    Example - Simple PURL:

        load("@package_metadata//purl:purl.bzl", "purl")

        my_purl = (purl.builder()
            .type("npm")
            .name("foobar")
            .version("12.3.1")
            .build())
        # Result: pkg:npm/foobar@12.3.1

    Example - Maven with namespace and qualifiers:

        load("@package_metadata//purl:purl.bzl", "purl")

        my_purl = (purl.builder()
            .type("maven")
            .namespace("org.apache.xmlgraphics")
            .name("batik-anim")
            .version("1.9.1")
            .add_qualifier("classifier", "sources")
            .add_qualifier("repository_url", "https://repo.spring.io/release")
            .build())
        # Result: pkg:maven/org.apache.xmlgraphics/batik-anim@1.9.1?classifier=sources&repository_url=https%3A%2F%2Frepo.spring.io%2Frelease

    Example - Golang with namespace and subpath:

        load("@package_metadata//purl:purl.bzl", "purl")

        my_purl = (purl.builder()
            .type("golang")
            .namespace("google.golang.org")
            .name("genproto")
            .version("abcdedf")
            .subpath("googleapis/api/annotations")
            .build())
        # Result: pkg:golang/google.golang.org/genproto@abcdedf#googleapis/api/annotations

    Returns:
        A builder object with chainable methods:

        - `type(type_name)`: Sets the package type (required). Must be lowercase ASCII.
        - `namespace(namespace)`: Sets the namespace (optional). String with segments separated by '/'.
        - `name(name)`: Sets the package name (required).
        - `version(version)`: Sets the package version (optional).
        - `add_qualifier(name, value)`: Adds a qualifier (optional, repeatable).
          Key must start with ASCII letter and contain only lowercase letters,
          numbers, '.', '-', '_'.
        - `subpath(subpath)`: Sets the subpath (optional). String with segments separated by '/'.
        - `build()`: Validates, normalizes, and constructs the final PURL string.
          Performs both general and type-specific validation and normalization.
          Fails if validation errors occur.
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
