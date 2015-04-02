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

set -x

mkdir -p /tmp/genroot
git clone $bare_root /tmp/genroot
(cd /tmp/genroot && go run /tmp/git-anchor.go -lang=$lang /tmp/tests/deps.json > /tmp/deps.sh)
(cd /tmp/genroot && go run /tmp/git-anchor.go -lang=$lang /tmp/tests/internal.json > /tmp/internal.sh)
rm -rf /tmp/genroot
chmod +x /tmp/deps.sh
chmod +x /tmp/internal.sh

error_expected=false
warning_expected=false

mkdir -p $root
cd $root

if [ $testname = "nothing" ]; then
	/tmp/tests/nothing.sh
elif [ $testname = "cloned" ]; then
	/tmp/tests/cloned.sh
elif [ $testname = "subtreed" ]; then
	/tmp/tests/subtreed.sh
elif [ $testname = "cloned_old" ]; then
	/tmp/tests/cloned_old.sh
	warning_expected=true
elif [ $testname = "subtreed_old" ]; then
	/tmp/tests/subtreed_old.sh
	warning_expected=true
elif [ $testname = "ingit" ]; then
	/tmp/tests/ingit.sh
	error_expected=true
elif [ $testname = "wrongfolder" ]; then
	/tmp/tests/subtreed.sh
	cd src
	error_expected=true
elif [ $testname = "notsquashed" ]; then
	/tmp/tests/justinternal.sh
	if [ $lang = "bash" ]; then
		/tmp/internal.sh
	else
		echo "language $lang not implemented"
		exit 1
	fi
	if [ ! $? = 0 ]; then
		exit 1
	fi
	cat $root/$path_subtree/hellosub.txt
	exit $?
else
	echo "testname $testname not implemented"
	exit 1
fi

error=false
if [ $lang = "bash" ]; then
	out=`/tmp/deps.sh`
	if [ ! $? = 0 ]; then
		error=true
	fi
else
	echo "language $lang not implemented"
	exit 1
fi

if [ $warning_expected = true ]; then
	if [[ ! $out = *"WARNING"* ]]; then
		echo "expected WARNING"
		exit 1
	fi
else
	if [[ $out = *"WARNING"* ]]; then
		echo "unexpected WARNING"
		exit 1
	fi
fi

if [ $error = true ]; then
	if [ $error_expected = false ]; then
		echo "unexpected error"
		exit 1
	fi
else
	if [ $error_expected = true ]; then
		echo "expected error"
		exit 1
	fi
	$root/$path_subtree/check.sh
	exit $?
fi


