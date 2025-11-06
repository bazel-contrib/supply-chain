# Examples of applications and interactions with licenses

load("@rules_cc//cc:cc_library.bzl", "cc_library")

package(
    default_visibility = ["//visibility:public"],
)

cc_library(
    name = "private_lib",
    srcs = [
        "private_lib.cc",
    ],
    hdrs = [
        "private_lib.h",
    ],
)

cc_library(
    name = "lib_with_third_party_deps",
    deps = [
        "//vendor/acme",
    ],
)
