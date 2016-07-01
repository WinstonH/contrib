# Copyright 2016 The Kubernetes Authors.
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

ARCH:=amd64
IMAGE_FORMAT:=vbox
DOCKER_IMAGE:=girishkalele/builders/packer-builder
TAG:=0.1
PACKERZIP:=packer_0.10.1_linux_$(ARCH).zip
PACKERTOOLS:=https://releases.hashicorp.com/packer/0.10.1/$(PACKERZIP)
K8S_WORKSPACE:=../../kubernetes

all: build-image

$(PACKERZIP):
	wget $(PACKERTOOLS)

packertools: $(PACKERZIP)
	unzip $(PACKERZIP)

container:
	docker build -t $(DOCKER_IMAGE):$(TAG) .

push: container
	docker push $(DOCKER_IMAGE):$(TAG)

build-kubernetes:
	cd $(K8S_WORKSPACE); make quick-release

build-iso: build-kubernetes
	@echo Building bootable ISO installer for Kubernetes

build-image: build-iso
	@echo Building $(IMAGE_FORMAT) VM image format for arch $(ARCH)


.PHONY: build-image
