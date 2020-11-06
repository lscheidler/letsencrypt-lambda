#!/usr/bin/env bash

if [ -f "letsencrypt-lambda" ] ; then
  OLD_SHA256SUM="sha256sum letsencrypt-lambda"
fi

OUTPUT=$(make build)
RVAL=$?

if [ "$RVAL" == "0" ] ; then
  NEW_SHA256SUM="sha256sum letsencrypt-lambda"

  # create/update zip, only if file is missing or sha256sum is different (source code changes)
  if [ ! -f "letsencrypt-lambda.zip" ] || [ "$OLD_SHA256SUM" != "$NEW_SHA256SUM" ]; then
    OUTPUT=$(make dist)
  fi

  echo "{\"filename\": \"$PWD/letsencrypt-lambda.zip\"}"
else
  echo "$OUTPUT"
  exit $RVAL
fi
