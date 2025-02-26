"""This attribute declares a license entity attached to a given target.
The attribute is a JSON with the following format:
- text_path: str - the path to the license file containing the text of the license
- full_name: str? - the human readable name of the license
- spdx_short_identifier: str? - the [SPDX short identifier](https://spdx.org/licenses/) of the license
- urls: str[] - the URLs pointing at the definition of this license.
The content of the license is stored as a separate file.
"""

visibility("public")

KIND = "build.bazel.well-known.license"
