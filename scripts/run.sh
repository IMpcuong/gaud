#!/bin/bash

# DEBUG: uncomment this line below to enable debug mode for this script.
# set -ex

# Global definitions:

declare -x env=$1 url=$2
declare -x app_path build_dir cur_dir=$(pwd)

# ---------- Utility functions ----------

function _detect_dir_to_run() {
  if [[ "${cur_dir}" == *"scripts"* ]]; then
    app_path=$(find .. -maxdepth 1 -type f -iname "*gad*" | xargs realpath)
    build_dir=$(dirname ${app_path})
    ${app_path} -d $url
    exit 0
  fi

  app_path=$(find . -maxdepth 1 -type f -iname "*gad*" | xargs realpath)
  build_dir=$(dirname ${app_path})
  ${app_path} -d $url
}

# ---------------------------------------

# Running process:

case $env in
  "docker")
    if [[ "" == $url ]]; then
      echo -ne "Error: Missing URI to download file\n"
      exit 0
    fi

    docker container run -it impcuong/gad:latest -d $url
    if (( "$?" != 0 )); then
      docker run -it --privileged -v $(pwd):/usr/local/bin/app/ impcuong/gad:latest -d $url
    fi
    declare -x container=$(docker ps -aq | head -n1)
    docker cp ${container}:/usr/local/bin/app/ .
    docker container rm -f ${container}
    ;;

  "local")
    if [[ "" == $url ]]; then
      echo -ne "Error: Missing URI to download file\n"
      exit 0
    fi
    _detect_dir_to_run
    ;;

  "")
    echo "Error: Unknown input argument to build application corresponded with specific environment"
    echo "Info: Valid options: 'local', 'docker'"
    exit 0
    ;;
esac