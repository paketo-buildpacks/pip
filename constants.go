package pip

// Pip is the name of the layer into which pip depdendency is installed.
const Pip = "pip"

// CPython is the name of the python runtime dependency provided by the CPython buildpack: https://github.com/paketo-buildpacks/cpython
const CPython = "cpython"

// DependencySHAKey is the name of the key in the pip layer TOML whose value is pip dependency's SHA256.
const DependencySHAKey = "dependency_sha"

// Priorities is a list of possible places where the buildpack could look for a
// specific version of Pip to install, ordered from highest priority to lowest.
var Priorities = []interface{}{"BP_PIP_VERSION"}
