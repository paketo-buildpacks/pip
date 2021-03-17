# Pip Rearchitecture

## Proposal

The existing pip buildpack should be rewritten and restructured to *only*
provide the `pip` dependency. The `pip install` logic should be factored out
into it's own buildpack.

## Motivation

In keeping with the overarching [Python Buildpack Rearchitecture
RFC](https://github.com/paketo-community/python/blob/main/rfcs/0001-restructure.md),
the Pip Buildpack should perform one task, which is installing the `pip`
dependency. This is part of the effort in Paketo Buildpacks to reduce the
responsibilities of each buildpack to make them easier to understand
and maintain.

## Implementation

The implementation details are outlined in [this
issue](https://github.com/paketo-community/pip/issues/82). Specifically, the
new Pip Buildpack will always `detect` and  will always `provide` `pip`. It
will be the responsibility of a downstream buildpack (such as a future Pip
Install buildpack) to `require` the `pip` dependency.

The new `provides`/`requires` contract will initially be:

* `pip`
  * provides: `pip`
  * requires: `cpython` OR {`python` + `requirements`} during `build`

The {`python` + `requirements`} requirement is included for
backwards-compatibility and will be removed towards the end of the full Python
rearchitecture.


The final `provides`/`requires` contract will be:

* `pip`
  * provides: `pip`
  * requires: `cpython` during `build`
