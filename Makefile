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

test:
	make install
	make build-testdocker
	make run-testdockers

install:
	go build git-anchor.go

build-testdocker:
	docker build -t deptestimage .

run-testdockers:
	docker rm deptestcontainer || true 
	docker run --env testname="nothing" --env lang=bash --rm=true -i -t --name deptestcontainer deptestimage
	docker run --env testname="cloned" --env lang=bash --rm=true -i -t --name deptestcontainer deptestimage
	docker run --env testname="subtreed" --env lang=bash --rm=true -i -t --name deptestcontainer deptestimage
	docker run --env testname="cloned_old" --env lang=bash --rm=true -i -t --name deptestcontainer deptestimage
	docker run --env testname="subtreed_old" --env lang=bash --rm=true -i -t --name deptestcontainer deptestimage
	docker run --env testname="ingit" --env lang=bash --rm=true -i -t --name deptestcontainer deptestimage
	docker run --env testname="wrongfolder" --env lang=bash --rm=true -i -t --name deptestcontainer deptestimage
	docker run --env testname="notsquashed" --env lang=bash --rm=true -i -t --name deptestcontainer deptestimage
