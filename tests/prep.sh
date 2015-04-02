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

#!/bin/bash
set -xe

#depends on dep
tmp_subtree=/tmp/subtree
mkdir -p $bare_subtree
(cd $bare_subtree && git init --bare)
mkdir -p $tmp_subtree
git clone $bare_subtree $tmp_subtree
(cd $tmp_subtree && mv /tmp/tests/check.sh . && chmod 777 check.sh && git add . && git commit -a -m 'checkdep' && git push origin master)
(cd $tmp_subtree && git rev-parse HEAD)
(rm -rf $tmp_subtree)

#depends on nothing
tmp_dep=/tmp/dep
mkdir -p $bare_dep
(cd $bare_dep && git init --bare)
mkdir -p $tmp_dep
git clone $bare_dep $tmp_dep
(cd $tmp_dep && echo 'hello' > hello.txt && git add . && git commit -a -m 'hello' && git push origin master)
(cd $tmp_dep && git rev-parse HEAD)
(rm -rf $tmp_dep)

#depends on dep and subtree
tmp_internal=/tmp/internal
mkdir -p $bare_internal
(cd $bare_internal && git init --bare)
mkdir -p $tmp_internal
git clone $bare_internal $tmp_internal
(cd $tmp_internal && echo 'hey' > hey.txt && git add . && git commit -a -m 'hey' && git push origin master)
(cd $tmp_internal && git rev-parse HEAD)
(rm -rf $tmp_internal)

tmp_old=/tmp/old
mkdir -p $bare_old
(cd $bare_old && git init --bare)
mkdir -p $tmp_old
git clone $bare_old $tmp_old
(cd $tmp_old && echo 'old' > old.txt && git add . && git commit -a -m 'old' && git push origin master)
(cd $tmp_old && git subtree add --prefix=$path_dep $bare_dep master --squash)
(cd $tmp_old && git subtree pull --prefix=$path_dep $bare_dep master --squash)
(cd $tmp_old && git subtree add --prefix=$path_subtree $bare_subtree master)
(cd $tmp_old && git subtree pull --prefix=$path_subtree $bare_subtree master)
(cd $tmp_old && git subtree add --prefix=$path_internal $bare_internal master)
(cd $tmp_old && git subtree pull --prefix=$path_internal $bare_internal master)
(cd $tmp_old && git push origin master)
(cd $tmp_old && git rev-parse HEAD)
(rm -rf $tmp_old )

mkdir -p $tmp_dep
git clone $bare_dep $tmp_dep
(cd $tmp_dep && echo 'hello2' > hello2.txt && git add . && git commit -a -m 'hello2' && git push origin master)
(cd $tmp_dep && git rev-parse HEAD)
(rm -rf $tmp_dep)

tmp_root=/tmp/root
mkdir -p $bare_root
(cd $bare_root && git init --bare)
mkdir -p $tmp_root
git clone $bare_root $tmp_root
(cd $tmp_root && echo 'rooty' > rooty.txt && git add . && git commit -a -m 'rooty' && git push origin master)
(cd $tmp_root && git subtree add --prefix=$path_dep $bare_dep master --squash)
(cd $tmp_root && git subtree pull --prefix=$path_dep $bare_dep master --squash)
(cd $tmp_root && git subtree add --prefix=$path_subtree $bare_subtree master)
(cd $tmp_root && git subtree pull --prefix=$path_subtree $bare_subtree master)
(cd $tmp_root && git subtree add --prefix=$path_internal $bare_internal master)
(cd $tmp_root && git subtree pull --prefix=$path_internal $bare_internal master)
(cd $tmp_root && git push origin master)
(cd $tmp_root && git rev-parse HEAD)

mkdir -p $tmp_subtree
git clone $bare_subtree $tmp_subtree
(cd $tmp_subtree && echo 'hello subtree' > hellosub.txt && git add . && git commit -a -m 'hello subtree' && git push origin master)
(cd $tmp_subtree && git rev-parse HEAD)
(rm -rf $tmp_subtree)

(cd $tmp_root && git subtree pull --prefix=$path_subtree $bare_subtree master)
(cd $tmp_root && git push origin master)
(cd $tmp_root && git rev-parse HEAD)
(rm -rf $tmp_root )


