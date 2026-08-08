#!/usr/bin/env python3

import sys
import tomllib

file_path = sys.argv[1]

with open(file_path, "rb") as f:
    data = tomllib.load(f)
    requires = data["build-system"]["requires"][0]
    _, constraints = requires.split(" ")
    print(constraints)
