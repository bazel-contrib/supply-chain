"""Normalization helpers for rule-declared toolchain involvement."""

def normalize_toolchain_type(label):
    """Return a toolchain label in apparent repository form."""
    value = str(label)
    if value.startswith("@@//"):
        return value[2:]
    if value.startswith("@@"):
        value = value[1:]
    if value.startswith("@") and not value.startswith("@@"):
        repository_end = value.find("//")
        if repository_end >= 0:
            repository = value[1:repository_end]
            # Module extension repositories (`module++extension+name`) are
            # distinct repositories that share the module prefix, so they must
            # keep their full name. Only module repositories (`module+` or
            # `module~version`) are reduced to the module name.
            if "++" not in repository:
                for separator in ["+", "~"]:
                    separator_index = repository.find(separator)
                    if separator_index >= 0:
                        repository = repository[:separator_index]
            value = "@" + repository + value[repository_end:]
    return value

