#!/bin/bash
# Copyright (c) 2017 Intel Corporation 
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

# List of files:
# 1. go files: *.go //
# 2. scripts: *.sh #
# 3. documents: *.md <!-- -->
# 4. Vagrantfile #
# 5. Dockerfile #
# 6. Python ''' '''

dirs="docs experiments integration_tests misc jupyter pkg scripts vagrant"


# $1 - should be comment file
# $2 - should be list of files
function add_header {
    local comment_file=${1}
    local list_of_files="${2}"

    for file in ${list_of_files}
    do
	if ! grep -q -e "Copyright" ${file}
	then
	    cat ${comment_file} ${file} > ${file}.tmp
	    mv ${file}.tmp ${file}
	fi
    done
}


##### SHELL ####
comment_file=/tmp/sh_license
files=`find ${dirs} -name "*.sh"`
sed 's/^/#/' < scripts/license.txt > ${comment_file}
echo >> ${comment_file}
add_header ${comment_file} "Makefile ${files}"

#### VAGRANTFILE ####
files=`find ${dirs} -name "Vagrantfile"`
add_header ${comment_file} "${files}"

#### DOCKERFILE ####
files=`find ${dirs} -name "Dockerfile"`
add_header ${comment_file} "Dockerfile ${files}"

rm -f ${comment_file}

#### GOLANG ####
files=`find ${dirs} -name "*.go"`
sed 's/^/\/\//' < scripts/license.txt > ${comment_file}
echo >> ${comment_file}

add_header ${comment_file} "${files}"
rm -f ${comment_file}


#### DOCS ####
files=`find ${dirs} -name "*.md"`
echo '<!--' > ${comment_file}
cat scripts/license.txt >> ${comment_file}
echo '-->' >> ${comment_file}
echo >> ${comment_file}

add_header ${comment_file} "README.md ${files}"
rm -f ${comment_file}

#### PYTHON ####
files=`find ${dirs} -name "*.py"`
echo '"""' > ${comment_file}
cat scripts/license.txt >> ${comment_file}
echo '"""' >> ${comment_file}
echo >> ${comment_file}

add_header ${comment_file} "${files}"
rm -f ${comment_file}

