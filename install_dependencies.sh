#!/bin/bash
dir=`dirname $0`
while read p; do
    go get -v ${p}
done < ${dir}/dependencies.txt