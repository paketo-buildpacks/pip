Note that compilation occurs on Jammy, but the result is not specific to Jammy.

Running compilation locally:

1. Build the build environment:
```shell
docker build \
  --tag pip-compilation-noarch \
  --file noarch.Dockerfile \
  .
```

2. Make a directory for the compiled output:
```shell
output_dir=$(mktemp -d)
```

3. Run compilation and use a volume mount to access the output directory:
```shell
docker run \
  --volume $output_dir:/tmp/compilation \
  pip-compilation-noarch \
  --outputDir /tmp/compilation \
  --target noarch \
  --version 22.2.2
```
