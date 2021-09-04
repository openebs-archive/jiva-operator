#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This file has been taken from https://github.com/kubernetes/code-generator/blob/HEAD/generate-groups.sh
# A small modification is made in this file at line number 56.

set -o errexit
set -o nounset
set -o pipefail

# generate-groups generates everything for a project with external types only, e.g. a project based
# on CustomResourceDefinitions.

#Usage: $(basename $0) <generators> <output-package> <apis-package> <groups-versions> ...
#
#  <generators>        the generators comma separated to run (deepcopy,defaulter,client,lister,informer) or "all".
#  <output-package>    the output package name (e.g. github.com/example/project/pkg/generated).
#  <apis-package>      the external types dir (e.g. github.com/example/api or github.com/example/project/pkg/apis).
#  <groups-versions>   the groups and their versions in the format "groupA:v1,v2 groupB:v1 groupC:v2", relative
#                      to <api-package>.
#  ...                 arbitrary flags passed to all generator binaries.
#
#
#Examples:
#  $(basename $0) all             github.com/example/project/pkg/client github.com/example/project/pkg/apis "foo:v1 bar:v1alpha1,v1beta1"
#  $(basename $0) deepcopy,client github.com/example/project/pkg/client github.com/example/project/pkg/apis "foo:v1 bar:v1alpha1,v1beta1"

(
  # To support running this script from anywhere, we have to first cd into this directory
  # so we can install the tools.
  #cd $(dirname "${0}")
  cd vendor/k8s.io/code-generator/ 
  go install ./cmd/{defaulter-gen,deepcopy-gen}
)

function codegen::join() { local IFS="$1"; shift; echo "$*"; }

module_name="github.com/openebs/jiva-operator"

# Generate deepcopy functions for all internalapis and external APIs
deepcopy_inputs=(
  pkg/apis/openebs/v1alpha1 \
)

gen-deepcopy() {
#  clean pkg/apis 'zz_generated.deepcopy.go'
  echo "Generating deepcopy methods..." >&2
  prefixed_inputs=( "${deepcopy_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )
  "${GOPATH}/bin/deepcopy-gen" \
    --go-header-file hack/custom-boilerplate.go.txt \
    --input-dirs "$joined" \
    --output-file-base zz_generated.deepcopy \
    --bounding-dirs "${module_name}"
#  for dir in "${deepcopy_inputs[@]}"; do
#    copyfiles "$dir" "zz_generated.deepcopy.go"
#  done
}

gen-deepcopy

