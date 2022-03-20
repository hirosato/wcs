#!/bin/sh
# Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

env=$1
echo "**********************************************"
echo "* Pushing Lambda '$env' "
echo "***********************************************"
if [ -z "$env" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh lbuild.sh <env>"
else
    aws lambda update-function-code --function-name wcs-$env --zip-file fileb://bin/main.zip
fi
