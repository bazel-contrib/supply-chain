<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares the package supplier attribute.

<a id="supplier"></a>

## supplier

<pre>
load("@package_metadata//attributes:supplier.bzl", "supplier")

supplier(*, <a href="#supplier-name">name</a>, <a href="#supplier-entity_name">entity_name</a>, <a href="#supplier-type">type</a>, <a href="#supplier-url">url</a>, <a href="#supplier-email">email</a>, <a href="#supplier-tags">tags</a>, <a href="#supplier-visibility">visibility</a>)
</pre>

Declares who distributes a package.

The supplier is the entity that distributes the package, which can differ
from the entity that originally created it. This is typically passed to
the `attributes` of a `package_metadata` target.

Usage:

```starlark
load("@package_metadata//attributes:supplier.bzl", "supplier")

supplier(
    name = "supplier",
    entity_name = "ACME Corp",
    url = "https://acme.example",
    email = "oss@acme.example",
)
```

Format: the `PackageAttributeInfo` kind is
`build.bazel.attribute.supplier`; its JSON object has `name` and `type`,
and includes `url` and `email` only when set. CycloneDX maps this to the
component supplier entity, always as an organization even when `type` is
`person`; URL and email are preserved. SPDX 2.3 maps it to
`PackageSupplier`, preserving type and email (in parentheses) but not URL.
SPDX 3.0.1 maps it to a `Person` or `Organization` referenced by the
package's `suppliedBy`, preserving email as an email external identifier
and URL as an `altWebPage` external reference.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="supplier-name"></a>name |  A unique name for this target.   |  none |
| <a id="supplier-entity_name"></a>entity_name |  Required. UTF-8 name of the package distributor.   |  none |
| <a id="supplier-type"></a>type |  `organization` or `person`; defaults to `organization`.   |  `"organization"` |
| <a id="supplier-url"></a>url |  Optional URL for the supplier.   |  `None` |
| <a id="supplier-email"></a>email |  Optional email address for the supplier.   |  `None` |
| <a id="supplier-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="supplier-visibility"></a>visibility |  The visibility of this target.   |  `None` |


