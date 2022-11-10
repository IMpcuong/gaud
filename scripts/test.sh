#!/bin/bash

declare -x app_path build_dir

declare cur_dir=$(pwd)
if [[ "${cur_dir}" == *"scripts"* ]]; then
  app_path=$(find .. -type f -iname "*gad*" | xargs realpath)
  build_dir=$(dirname ${app_path})
  go test -v ${build_dir}/...
  exit 0
fi

app_path=$(find . -type f -iname "*gad*" | xargs realpath)
build_dir=$(dirname ${app_path})
go test -v ${build_dir}/...
