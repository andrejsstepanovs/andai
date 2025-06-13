#!/bin/bash

# This script is combining all md files from the project into a single file named "singlefile.md"
output_file="docs/singlefile.md"

# Clear the output file if it exists
if [ -f "$output_file" ]; then
    rm "$output_file"
fi

find docs -type f -name "*.md" | while read -r file; do
    if [ "$file" != "./$output_file" ]; then
        echo "Processing $file"
        echo -e "\n# $(basename "$file")\n" >> "$output_file"
        cat "$file" >> "$output_file"
        echo -e "\n\n---\n\n" >> "$output_file"
    fi
done

EXTRA_FILES=("swagger.andai.project.yaml" "README.md")

for extra_file in "${EXTRA_FILES[@]}"; do
    if [ -f "$extra_file" ]; then
        echo "Adding $extra_file to the top of $output_file"
        echo -e "\n# $(basename "$extra_file")\n" >> "$output_file"
        cat "$extra_file" >> "$output_file"
        echo -e "\n\n---\n\n" >> "$output_file"
    fi
done
