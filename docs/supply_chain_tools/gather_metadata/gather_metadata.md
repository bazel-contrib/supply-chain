<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Rules and macros for collecting LicenseInfo providers.

<a id="metadata_info_to_json"></a>

## metadata_info_to_json

<pre>
load("@supply_chain_tools//gather_metadata:gather_metadata.bzl", "metadata_info_to_json")

metadata_info_to_json(<a href="#metadata_info_to_json-metadata_info">metadata_info</a>)
</pre>

Render a single LicenseInfo provider to JSON

**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="metadata_info_to_json-metadata_info"></a>metadata_info |  A LicenseInfo.   |  none |

**RETURNS**

[(str)] list of LicenseInfo values rendered as JSON.


<a id="write_metadata_info"></a>

## write_metadata_info

<pre>
load("@supply_chain_tools//gather_metadata:gather_metadata.bzl", "write_metadata_info")

write_metadata_info(<a href="#write_metadata_info-ctx">ctx</a>, <a href="#write_metadata_info-deps">deps</a>, <a href="#write_metadata_info-json_out">json_out</a>)
</pre>

Writes TransitiveMetadataInfo providers for a set of targets as JSON.

TODO(aiuto): Document JSON schema. But it is under development, so the current
best place to look is at tests/hello_licenses.golden.

Usage:
  write_metadata_info must be called from a rule implementation, where the
  rule has run the gather_metadata_info aspect on its deps to
  collect the transitive closure of LicenseInfo providers into a
  LicenseInfo provider.

  foo = rule(
    implementation = _foo_impl,
    attrs = {
       "deps": attr.label_list(aspects = [gather_metadata_info])
    }
  )

  def _foo_impl(ctx):
    ...
    out = ctx.actions.declare_file("%s_licenses.json" % ctx.label.name)
    write_metadata_info(ctx, ctx.attr.deps, metadata_file)


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="write_metadata_info-ctx"></a>ctx |  context of the caller   |  none |
| <a id="write_metadata_info-deps"></a>deps |  a list of deps which should have TransitiveMetadataInfo providers. This requires that you have run the gather_metadata_info aspect over them   |  none |
| <a id="write_metadata_info-json_out"></a>json_out |  output handle to write the JSON info   |  none |


<a id="gather_metadata_info"></a>

## gather_metadata_info

<pre>
load("@supply_chain_tools//gather_metadata:gather_metadata.bzl", "gather_metadata_info")

gather_metadata_info()
</pre>

Collects metadata providers into a single TransitiveMetadataInfo provider.

**ASPECT ATTRIBUTES**



**ATTRIBUTES**



<a id="gather_metadata_info_and_write"></a>

## gather_metadata_info_and_write

<pre>
load("@supply_chain_tools//gather_metadata:gather_metadata.bzl", "gather_metadata_info_and_write")

gather_metadata_info_and_write()
</pre>

Collects TransitiveMetadataInfo providers and writes JSON representation to a file.

Usage:
  bazel build //some:target           --aspects=@rules_license//rules_gathering:gather_metadata.bzl%gather_metadata_info_and_write
      --output_groups=licenses

**ASPECT ATTRIBUTES**



**ATTRIBUTES**



