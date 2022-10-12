Note that compilation occurs on Jammy, but the result is not specific to Jammy.

Running compilation locally:

1. Build the build environment:
```shell
docker build --tag compilation-noarch --file noarch.Dockerfile .
```

2. Make the output directory:
```shell
output_dir=$(mktemp -d)
```

3. Run compilation and use a volume mount to access it:
```shell
docker run --volume $output_dir:/tmp/compilation compilation-noarch --outputDir /tmp/compilation --target noarch --version 22.2.2
```
