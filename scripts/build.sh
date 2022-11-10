#!/bin/bash

# DEBUG: uncomment this line below to enable debug mode for this script.
# set -ex

# Global definitions:

declare -x env=$1

declare -x app_path build_dir env_path
declare -x cur_dir=$(pwd)

# ---------- Utility functions ----------

function _exist() {
  which $env >/dev/null 2>&1
  if (( "$?" != 0 )); then
    # NOTE:
    # - Uppercase the required application's name, using 2 simple methods:
    #   + Solution1: the `tr` command, noting that `'[:lower:]'` equals to `[[:lower:]]`.
    #   + Solution2: variable expansion such as `${env^}`.
    env="$(tr '[:lower:]' '[:upper:]' <<< ${env:0:1})${env:1}"
    echo "Err: ${env^} was not installed. Please install $env"
  fi
}

function _detect_dir_to_build() {
  if [[ "${cur_dir}" == *"scripts"* ]]; then
    app_path=$(find ../ -maxdepth 1 -type f -iname "*gad*" | xargs realpath)
    build_dir=$(dirname ${app_path})
    go build -v -o ${build_dir}/gad.exe ${build_dir}/...
    exit 0
  fi

  app_path=$(find ./ -maxdepth 1 -type f -iname "*gad*" | xargs realpath)
  build_dir=$(dirname ${app_path})
  go build -v -o ${build_dir}/gad.exe ${build_dir}/...
}

function _export_proxies() {
  export http_proxy="${proxy_host}:${proxy_port}"
  export https_proxy="${proxy_host}:${proxy_port}"
}

function _detect_env_file() {
  if [[ "${cur_dir}" == *"scripts"* ]]; then
    env_path=$(find ../ -type f -a -regextype "egrep" -regex ".*\.env" | xargs realpath)
    source ${env_path} && _export_proxies
    exit 0
  fi

  env_path=$(find ./ -type f -a -regextype "egrep" -regex ".*\.env" | xargs realpath)
  source ${env_path} && _export_proxies
}

# ---------------------------------------

# Building process:

_detect_env_file
case $env in
  "docker")
    if ! _exist; then exit 0; fi
    docker images -a | grep gad
    if (( "$?" != 0 )); then
      docker build --tag impcuong/gad:latest .
    else
      docker images -a | \
        grep gad | \
        awk '{ print $3 }' | \
        xargs docker rmi -f
      docker build --tag impcuong/gad:latest .
    fi
    ;;

  "local")
    if [[ "local" == $env ]]; then env="go"; fi
    if ! _exist; then exit 0; fi
    _detect_dir_to_build
    ;;

  "")
    echo "Error: Unknown input argument to build application corresponded with specific environment"
    echo "Info: Valid options: 'local', 'docker'"
    exit 0
esac
