#!/usr/bin/make -f

# Copyright (c) 2022  The Go-Enjin Authors
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

#: uncomment to echo instead of execute
#CMD=echo

SHELL = /bin/bash

BE_PATH ?= ../../../../be

BUILD_TAGS = sass
EXTRA_PKGS =

GOLANG ?= 1.17.7
NODEJS ?=

GO_ENJIN_PKG = github.com/go-enjin/be

.PHONY: all help enjenv golang tidy local unlocal be-update build

help:
	@echo "usage: make <help|tidy|local|unlocal|be-update|build>"

define _be_local_path =
$(shell \
	if [ "${BE_LOCAL_PATH}" != "" -a -d "${BE_LOCAL_PATH}" ]; then \
		echo "${BE_LOCAL_PATH}"; \
	elif [ "${GOPATH}" != "" ]; then \
		if [ -d "${GOPATH}/src/${GO_ENJIN_PKG}" ]; then \
			echo "${GOPATH}/src/${GO_ENJIN_PKG}"; \
		fi; \
	fi)
endef

define _be_version =
$(shell ${ENJENV_EXE} git-tag --untagged v0.0.0)
endef

define _be_release =
$(shell ${ENJENV_EXE} rel-ver)
endef

define _enjenv_bin =
$(shell \
	if [ "${ENJENV_BIN}" != "" -a -x "${ENJENV_BIN}" ]; then \
		echo "${ENJENV_BIN}"; \
	else \
		if [ "$(1)" != "" ]; then \
			echo "$(1)"; \
		elif [ -d .bin -a -x .bin/enjenv ]; then \
			echo "${PWD}/.bin/enjenv"; \
		fi; \
	fi)
endef

define _enjenv_path =
$(shell \
	if [ -x "${ENJENV_EXE}" ]; then \
		${ENJENV_EXE}; \
	elif [ -d "./${ENJENV_DIR}" ]; then \
		echo "${PWD}/${ENJENV_DIR}"; \
	fi)
endef

define _build_tags =
$(shell if [ "${BUILD_TAGS}" != "" ]; then echo "-tags ${BUILD_TAGS}"; fi)
endef

BE_PATH ?= $(call _be_local_path)

ENJENV_EXE ?= $(call _enjenv_bin,$(shell which enjenv))
ENJENV_URL ?= https://github.com/go-enjin/enjenv-heroku-buildpack/raw/trunk/bin/enjenv
ENJENV_PKG ?= github.com/go-enjin/enjenv/cmd/enjenv@latest
ENJENV_DIR_NAME ?= .enjenv
ENJENV_DIR ?= ${ENJENV_DIR_NAME}
ENJENV_PATH ?= $(call _enjenv_path)

VERSION ?= $(call _be_version)
RELEASE ?= $(call _be_release)

enjenv: gobin=$(shell which go)
enjenv:
	@if [ "${ENJENV_EXE}" = "" ]; then \
		if [ "${gobin}" = "" ]; then \
			echo "# downloading enjenv..."; \
			wget -q -c ${ENJENV_URL}; \
			chmod +x ./enjenv; \
			echo "# installing enjenv..."; \
			if [ "${GOPATH}" = "" ]; then \
				mkdir -v .bin; \
				mv -v ./enjenv ./.bin/; \
			else \
				mv -v ./enjenv ${GOPATH}/bin/enjenv; \
			fi; \
		else \
			echo "# go install enjenv..."; \
			go install ${ENJENV_PKG}; \
		fi; \
	fi

golang: enjenv
	@if [ "${ENJENV_PATH}" != "" ]; then \
		if [ ! -d "${ENJENV_PATH}/golang" ]; then \
			if [ "${GOLANG}" != "" ]; then \
				${CMD} ${ENJENV_EXE} golang init --golang "${GOLANG}"; \
				${CMD} ${ENJENV_EXE} write-scripts; \
			else \
				${CMD} ${ENJENV_EXE} golang init; \
				${CMD} ${ENJENV_EXE} write-scripts; \
			fi; \
		fi; \
		if [ "${NODEJS}" != "" ]; then \
			if [ ! -d "${ENJENV_PATH}/nodejs" ]; then \
				${CMD} ${ENJENV_EXE} nodejs init --nodejs "${NODEJS}"; \
				${CMD} ${ENJENV_EXE} write-scripts; \
			fi; \
		fi; \
	else \
		echo "# missing enjenv path"; \
		false; \
	fi

tidy: golang
	@echo "# go mod tidy -go=1.16 && go mod tidy -go=1.17"
	@source "${ENJENV_PATH}/activate" \
		&& ${CMD} go mod tidy -go=1.16 \
		&& ${CMD} go mod tidy -go=1.17

local: enjenv
	@if [ "${BE_PATH}" = "" ]; then \
		echo "missing BE_PATH"; \
		false; \
	fi
	@echo "# localizing ${GO_ENJIN_PKG}"
	@${CMD} ${ENJENV_EXE} go-local "${BE_PATH}"

unlocal: enjenv
	@echo "# restoring ${GO_ENJIN_PKG}"
	@${CMD} ${ENJENV_EXE} go-unlocal

be-update: golang
	@echo "# go get -u ${GO_ENJIN_PKG} ${EXTRA_PKGS}"
	@source "${ENJENV_PATH}/activate" \
		&& ${CMD} GOPROXY=direct go get -u \
			$(call _build_tags) \
			${GO_ENJIN_PKG} \
			${EXTRA_PKGS}

build: golang
	@source "${ENJENV_PATH}/activate" \
		&& ${CMD} go build -v $(call _build_tags)
