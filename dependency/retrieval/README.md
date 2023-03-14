# Dependency Retrieval

## Running locally

Run the following command:

```
go run main.go \
  --buildpack-toml-path ../../buildpack.toml \
  --output /path/to/retrieved.json
```

Example output (abbreviated for clarity):

```
Found 123 versions of pip from upstream
[
  "23.0.1", "23.0.0", [...],  "0.2.0"
]
Found 123 versions of pip for constraint *
[
  "23.0.1", "23.0.0", [...],  "0.2.0"
]
Found 2 versions of pip newer than '22.3.1' for constraint *, after limiting for 2 patches
[
  "23.0.1", "23.0.0"
]
Found 2 versions of pip as new versions
[
  "23.0.1", "23.0.0"
]
Generating metadata for 23.0.1, with targets [noarch]
Generating metadata for 23.0.0, with targets [noarch]
Wrote metadata to /path/to/retrieved.json

```
