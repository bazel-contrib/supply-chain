"""Provider for per-target direct dependency linkage classifications.

Return this from a rule when the way it links a dependency depends on how the
target is configured, for example a flag choosing a static or a dynamic library.
The SBOM tooling then reports each dependency as `static_link` or `dynamic_link`.
"""

visibility("public")

PackageLinkingInfo = provider(
    doc = """
Classifies direct dependency targets by how the providing target links them.
""".strip(),
    fields = {
        "static_linked": """
List of Labels for direct dependency targets this target links statically or
embeds. These must be labels of direct dependencies, not transitive targets.
""".strip(),
        "dynamic_linked": """
List of Labels for direct dependency targets this target links dynamically.
These must be labels of direct dependencies, not transitive targets.
""".strip(),
    },
)
