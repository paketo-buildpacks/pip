package pip

// Pip is the name of the layer into which pip dependency is installed.
const Pip = "pip"

const PipSrc = "pip-source"

// CPython is the name of the python runtime dependency provided by the CPython buildpack: https://github.com/paketo-buildpacks/cpython
const CPython = "cpython"

// DependencyChecksumKey is the name of the key in the pip layer TOML whose value is pip dependency's SHA256.
const DependencyChecksumKey = "dependency_checksum"

// Priorities is a list of possible places where the buildpack could look for a
// specific version of Pip to install, ordered from highest to lowest priority.
var Priorities = []interface{}{"BP_PIP_VERSION"}
