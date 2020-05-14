# PIP Cloud Native Buildpack

## Integration

The Pip CNB provides pip as a dependency. Downstream buildpacks can require the nginx
dependency by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the Pip dependency is "pip". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "pip"

  # The version of the Pip dependency is not required. In the case it
  # is not specified, the buildpack will provide the default version, which can
  # be seen in the buildpack.toml file.
  # If you wish to request a specific version, the buildpack supports
  # specifying a semver constraint in the form of "20.*", "20.0.*", or even
  # "20.0.2".
  version = "20.0.2"

  # The Pip buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the Pip
    # dependency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to run Pip
    # during its build process, this flag should be set to true.
    build = true

    # Setting the launch flag to true will ensure that the Pip
    # dependency is available on the $PATH for the running application. If you are
    # writing an application that needs to run Pip at runtime, this flag should
    # be set to true.
    launch = true
```

## Usage

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.
