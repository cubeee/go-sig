#!/bin/sh
while read dep; do gb vendor fetch "$dep"; done < dependencies.txt
