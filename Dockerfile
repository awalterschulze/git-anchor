# Copyright 2015 Vastech SA (PTY) LTD
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

FROM centos:centos7

RUN yum -y install git wget mercurial tar
RUN yum -y group install "Development Tools"  || true
RUN yum -y install golang

RUN git config --global user.email "me@example.com"
RUN git config --global user.name "My Name"

ADD . /tmp/
RUN rm -rf /tmp/.git || true
RUN chmod 777 -R ./tmp/

ENV root /tmp/workspace
ENV path_dep src/dep
ENV path_subtree src/subtree
ENV path_internal src/internal
ENV bare_subtree /subtree.git
ENV bare_dep /dep.git
ENV bare_internal /internal.git
ENV bare_old /old.git
ENV bare_root /root.git

RUN /tmp/tests/prep.sh

ENTRYPOINT /tmp/tests/tests.sh
