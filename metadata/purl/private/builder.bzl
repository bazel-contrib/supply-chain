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

def _build(
        self,
        type = None,
        namespace = None,
        name = None,
        version = None,
        qualifiers = {},
        subpath = None):
    validate(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )

    purl = normalize(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )

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
    components.append(purl.name)

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
            if key == "checksum":
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

    return "".join(components)

def builder():
    fields = {}
    self = struct(
        type = lambda type_name: _type(self, fields, type_name),
        namespace = lambda namespace: _namespace(self, fields, namespace),
        name = lambda name: _name(self, fields, name),
        version = lambda version: _version(self, fields, version),
        add_qualifier = lambda name, value: _add_qualifier(self, fields, name, value),
        subpath = lambda subpath: _subpath(self, fields, subpath),
        build = lambda: _build(self, **fields),
    )
    return self
