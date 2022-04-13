#!/usr/bin/env bash

# Check gofmt
echo "==> Checking for unchecked errors..."

if ! which errcheck > /dev/null; then
    echo "==> Installing errcheck..."
    go install github.com/kisielk/errcheck@latest
fi

err_files=$(errcheck -ignoretests \
                     -ignore 'bytes:.*' \
                     -ignore 'io:Close|Write' \
                     $(go list ./cmd/...| grep -v /vendor/))

if [[ -n ${err_files} ]]; then
    echo 'Unchecked errors found in the following places:'
    echo "${err_files}"
    echo "Please handle returned errors. You can check directly with \`make errcheck\`"
    exit 1
fi

exit 0