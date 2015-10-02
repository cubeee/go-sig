#!/bin/sh
while read dep; do gb vendor fetch "$dep" 2>&1 || true; done < dependencies.txt
