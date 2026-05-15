"""Normalization for MLflow PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/mlflow-definition.json
"""

visibility([
    "//purl/private/normalization/...",
])

def normalize_mlflow(components):
    repository_url = components["qualifiers"].get("repository_url", "") if components["qualifiers"] else ""
    if "azuredatabricks.net" in repository_url:
        components["name"] = components["name"].lower() if components["name"] else components["name"]
    return components
