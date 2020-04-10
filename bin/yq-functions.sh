#!/bin/bash
#set -xe

# make a filtering by arbitrary fields and return only matching yamls as yaml array
# args: $1 takes a string with bash logical expression to filter yaml (if true)
# e.g. '[ $(val kind) == "Secret" ] && [ $(val metadata.name) == "node1-bmc-secret" ]'
# output: prints matched yamls to stdout
function yq::filter() {
	local yamls=$(cat)
	
	local i=0
	local yamls_len=$(echo "$yamls" | yq r -d* - * | wc -l)

	function val() {
		echo "$yamls" | yq r -d$i - $1
	}
	
	local flag=0
	while [ $i != $yamls_len ]; do
		if eval $1; then
			[ $flag == 0 ] || echo '---' && flag=1
			val
		fi
		i=$(($i+1))
	done
}

#yq::filter "$1"
