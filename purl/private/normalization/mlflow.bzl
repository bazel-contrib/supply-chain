"""Normalization for MLflow PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/mlflow-definition.json
"""

visibility([
    "//private/normalization/...",
])

def normalize_mlflow(components):
    # While these are not described in the normalization section of the spec, they are validated as such in the
    # type specific tests
    repository_url = components["qualifiers"].get("repository_url", "") if components["qualifiers"] else ""
    if "azuredatabricks.net" in repository_url:
        components["name"] = components["name"].lower() if components["name"] else components["name"]
    return components
