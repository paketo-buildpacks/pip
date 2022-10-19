To test locally:

```shell
# assume $output_dir is the output from the compilation step, with a tarball and a checksum in it

docker run -it \
  --volume $output_dir:/tmp/output_dir \
  --volume $PWD:/tmp/test \
  ubuntu:jammy \
  bash

# Now on the container
# This is not required on Github Actions Virtual Environments
# https://github.com/actions/runner-images/blob/main/images/linux/Ubuntu2004-Readme.md
apt-get update && apt-get install python3 -y

# Passing
$ /tmp/test/test.sh \
  --tarballPath /tmp/output_dir/pip_22.2.2_noarch_ff717ff0.tgz \
  --expectedVersion 22.2.2
tarballPath=/tmp/output_dir/pip_22.2.2_noarch_ff717ff0.tgz
expectedVersion=22.2.2
All tests passed!

# Failing
$ /tmp/test/test.sh \
  --tarballPath /tmp/output_dir/pip_22.2.2_noarch_ff717ff0.tgz \
  --expectedVersion 999.999.999
tarballPath=/tmp/output_dir/pip_22.2.2_noarch_ff717ff0.tgz
expectedVersion=999.999.999
Version 22.2.2 does not match expected version 999.999.999
```