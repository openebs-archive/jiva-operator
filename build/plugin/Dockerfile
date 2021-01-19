# Copyright 2019-2020 The OpenEBS Authors
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

FROM ubuntu:18.04
RUN apt-get update; exit 0
RUN apt-get -y install rsyslog xfsprogs curl
RUN apt-get clean && rm -rf /var/lib/apt/lists/*

COPY build/bin/jiva-csi /usr/local/bin/

ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL
ARG ARCH

LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.name="jiva-csi"
LABEL org.label-schema.description="OpenEBS Jiva CSI Plugin"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL
LABEL org.label-schema.arch=$ARCH

ENTRYPOINT ["/usr/local/bin/jiva-csi"]
