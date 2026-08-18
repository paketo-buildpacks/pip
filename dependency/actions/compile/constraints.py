import sys
from subprocess import check_call

import tomllib

file_path = sys.argv[1]

with open(file_path, "rb") as f:
    data = tomllib.load(f)
    for entry in data["build-system"]["requires"]:
        name, constraints = entry.split(" ")
        check_call(
            [
                "pip3",
                "--cache-dir=/tmp-cache/",
                "download",
                "--no-binary",
                ":all:",
                f"{name}{constraints}",
            ]
        )
