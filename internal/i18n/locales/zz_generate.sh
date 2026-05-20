#!/bin/bash

set -euo pipefail

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

langs=$(find locales -type f -name "*.yaml" -exec basename {} .yaml \;)
keys=$(grep -oP '(?<=Key = ").*(?=")' keys.go)

# Set queue to langs if no args are provided
# Otherwise, set queue to the items that are both in args and langs
if [ "$#" -eq 0 ]; then
    queue="$langs"
else
    queue=""
    for arg in "$@"; do
        if [[ " $langs " == *" $arg "* ]]; then
            queue="$queue $arg"
        else
            echo "Warning: Language '$arg' is not in the list of available languages. Skipping."
        fi
    done
fi

for lang in $langs; do
    for key in $keys; do
        if [ ! -z "$(yq e "select(.[] | .id == \"$key\") | length" locales/$lang.yaml)" ]; then
            yq e ".[] | select(.id == \"$key\") | [.]" locales/$lang.yaml >> "$TMP_DIR/$lang.yaml"
        else
            echo "- id: $key" >> "$TMP_DIR/$lang.yaml"
            echo "  # TODO: translate" >> "$TMP_DIR/$lang.yaml"
            echo "  other: $key" >> "$TMP_DIR/$lang.yaml"
        fi
    done
    mv "$TMP_DIR/$lang.yaml" locales/$lang.yaml
done
