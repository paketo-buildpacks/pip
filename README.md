# PIP Cloud Native Buildpack
The Paketo Pip Buildpack is a Cloud Native Buildpack that installs pip into a
layer and places it on the `PATH`.

The buildpack is published for consumption at `gcr.io/paketo-community/pip` and
`paketocommunity/pip`.

## Behavior
This buildpack always participates.

The buildpack will do the following:
* At build time:
  - Contributes the `pip` binary to a layer
  - Prepends the `pip` layer to the `PYTHONPATH`
  - Adds the newly installed pip location to `PATH`
* At run time:
  - Does nothing

## Configuration
| Environment Variable | Description
| -------------------- | -----------
| `$BP_PIP_VERSION` | Configure the version of pip to install. Buildpack releases (and the pip versions for each release) can be found [here](https://github.com/paketo-community/pip/releases).

## Integration

The Pip CNB provides pip as a dependency. Downstream buildpacks can require the pip
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
  # is not specified, the buildpack will select the latest supported version in
  # the buildpack.toml.
  # If you wish to request a specific version, the buildpack supports
  # specifying a semver constraint in the form of "21.*", "21.0.*", or even
  # "21.0.1".
  version = "21.0.1"

  # The Pip buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the Pip dependency is
    # available on the $PATH, and the $PYTHONPATH contains the path to pip for
    # subsequent buildpacks during their build phase. If you are writing a
    # buildpack that needs to run Pip during its build process, this flag should
    # be set to true.
    build = true

    # Setting the launch flag to true will ensure that the Pip
    # dependency is available on the $PATH, and the $PYTHONPATH contains the
    # path to pip for the running application. If you are writing an
    # application that needs to run Pip at runtime, this flag should be set to
    # true.
    launch = true
```

## Usage

To package this buildpack for consumption:
```
$ ./scripts/package.sh --version x.x.x
```
This will create a `buildpackage.cnb` file under the build directory which you
can use to build your app as follows: `pack build <app-name> -p <path-to-app> -b
build/buildpackage.cnb -b <other-buildpacks..>`.

To run the unit and integration tests for this buildpack:
```
$ ./scripts/unit.sh && ./scripts/integration.sh
```
