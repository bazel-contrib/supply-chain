<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares the package originator attribute.

<a id="originator"></a>

## originator

<pre>
load("@package_metadata//attributes:originator.bzl", "originator")

originator(*, <a href="#originator-name">name</a>, <a href="#originator-entity_name">entity_name</a>, <a href="#originator-type">type</a>, <a href="#originator-url">url</a>, <a href="#originator-email">email</a>, <a href="#originator-tags">tags</a>, <a href="#originator-visibility">visibility</a>)
</pre>

Declares who originally created a package.

The originator is the entity that originally created the package; it may
differ from the supplier that distributes it. This is typically passed to
the `attributes` of a `package_metadata` target.

Usage:

```starlark
load("@package_metadata//attributes:originator.bzl", "originator")

originator(
    name = "originator",
    entity_name = "A. Creator",
    type = "person",
    email = "creator@example",
)
```

Format: the `PackageAttributeInfo` kind is
`build.bazel.attribute.originator`; its JSON object has `name` and `type`,
and includes `url` and `email` only when set. CycloneDX maps this to the
component publisher string, so type, URL, and email are not represented.
SPDX 2.3 maps it to `PackageOriginator`, preserving type and email (in
parentheses) but not URL. SPDX 3.0.1 maps it to a `Person` or
`Organization` referenced by the package's `originatedBy`, preserving
email as an email external identifier and URL as an `altWebPage` external
reference.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="originator-name"></a>name |  A unique name for this target.   |  none |
| <a id="originator-entity_name"></a>entity_name |  Required. UTF-8 name of the original creator.   |  none |
| <a id="originator-type"></a>type |  `organization` or `person`; defaults to `organization`.   |  `"organization"` |
| <a id="originator-url"></a>url |  Optional URL for the originator.   |  `None` |
| <a id="originator-email"></a>email |  Optional email address for the originator.   |  `None` |
| <a id="originator-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="originator-visibility"></a>visibility |  The visibility of this target.   |  `None` |


