#!/bin/bash

set -euo pipefail
shopt -s inherit_errexit

parent_dir="$(cd "$(dirname "$0")" && pwd)"

extract_tarball() {
  rm -rf pip
  mkdir pip
  tar --extract --file "${1}" \
    --directory pip
}

check_version() {
  expected_version="$1"

  if [[ "${expected_version}" =~ [0-9]+.[0-9]+.0 ]]; then
    new_expected_version="$(echo "${expected_version}" | cut -d '.' -f1,2)"
    echo "expected version ${expected_version} has '0' for patch - using 'Major.Minor' format instead (i.e.: '${new_expected_version}')"
    expected_version="${new_expected_version}"
  fi

  actual_version="$(grep -e "^Version:" pip/PKG-INFO | awk -F ': ' '{print $2}')"
  if [[ "${actual_version}" != "${expected_version}" ]]; then
    echo "Version ${actual_version} does not match expected version ${expected_version}"
    exit 1
  fi
}

main() {
  local tarballPath expectedVersion
  tarballPath=""
  expectedVersion=""

  while [ "${#}" != 0 ]; do
    case "${1}" in
      --tarballPath)
        tarballPath="${2}"
        shift 2
        ;;

      --expectedVersion)
        expectedVersion="${2}"
        shift 2
        ;;

      "")
        shift
        ;;

      *)
        echo "unknown argument \"${1}\""
        exit 1
    esac
  done

  if [[ "${tarballPath}" == "" ]]; then
    echo "--tarballPath is required"
    exit 1
  fi

  if [[ "${expectedVersion}" == "" ]]; then
    echo "--expectedVersion is required"
    exit 1
  fi

  echo "tarballPath=${tarballPath}"
  echo "expectedVersion=${expectedVersion}"

  extract_tarball "${tarballPath}"
  check_version "${expectedVersion}"

  echo "All tests passed!"
}

main "$@"
