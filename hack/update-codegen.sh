#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${REPO_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

KNATIVE_CODEGEN_PKG=${KNATIVE_CODEGEN_PKG:-$(cd ${REPO_ROOT}; ls -d -1 ./vendor/knative.dev/pkg 2>/dev/null || echo ../pkg)}
GO_MODULE=my.dev/solar-system

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
  $GO_MODULE/pkg/client $GO_MODULE/pkg/apis \
  "solar:v1alpha1" \
  --go-header-file ${REPO_ROOT}/hack/boilerplate/boilerplate.go.txt

# Knative Injection
${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh "injection" \
  $GO_MODULE/pkg/client $GO_MODULE/pkg/apis \
  "solar:v1alpha1" \
  --go-header-file ${REPO_ROOT}/hack/boilerplate/boilerplate.go.txt

# Depends on generate-groups.sh to install bin/deepcopy-gen
${GOPATH}/bin/deepcopy-gen \
  -O zz_generated.deepcopy \
  --go-header-file ${REPO_ROOT}/hack/boilerplate/boilerplate.go.txt \
  -i $GO_MODULE/pkg/apis/solar/v1alpha1

# Make sure our dependencies are up-to-date
#${REPO_ROOT}/hack/update-deps.sh
