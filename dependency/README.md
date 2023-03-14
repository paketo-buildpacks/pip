# Dependency

Pre-compiled distributions of Pip are provided for all linux stacks.

This directory contains scripts and GitHub Actions to facilitate the following:
* Identifying when there is a new version of Pip available
* Compiling Pip for all supported stacks (i.e. `noarch`)

## Running locally

Running the steps locally can be useful for iterating on the compilation process
(e.g. changing compilation options) as well as debugging.

### Retrieval

Retrieve latest versions with:

```
cd ./retrieval

go run main.go \
  --buildpack-toml-path ../../buildpack.toml \
  --output /path/to/retrieved.json
```

See [retrieval/README.md](retrieval/README.md) for more details.

### Compilation

To compile:

```
docker build \
  --tag pip-compilation-noarch \
  --file ./actions/compile/noarch.Dockerfile \
  ./actions/compile

output_dir=$(mktemp -d)

docker run \
  --volume $output_dir:/tmp/compilation \
  pip-compilation-noarch \
    --outputDir /tmp/compilation \
    --target noarch \
    --version 23.0.1
```

See [actions/compile/README.md](actions/compile/README.md) for more details.
