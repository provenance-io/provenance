#!/bin/bash

LIB_PATH=${LIB_PATH:-/usr/local/lib}
LIB_RDKAFKA=librdkafka.so

# Check if we have /usr/local/lib/librdkafka.so
if [ -f "$LIB_PATH/$LIB_RDKAFKA" ]; then
    :
else
    echo "Missing $dependency. Recommended to run: make librdkafka"
    exit 1
fi

# Check if we have the env varibale set
if ! tr ':' '\n' <<< "$LD_LIBRARY_PATH" | grep -xFq "$LIB_PATH"; then
    echo "LD_LIBRARY_PATH is missing ${LIB_PATH}. Recommended to run: export LD_LIBRARY_PATH=\${LD_LIBRARY_PATH:+\$LD_LIBRARY_PATH:}$LIB_PATH"
    exit 1
else
    echo "LD_LIBRARY_PATH is already set."
fi
exit 0