# !/bin/bash

while true; do
    OUTPUT=$(PROJECT=streamlit make work 2>&1)
    echo "$OUTPUT"  # Optionally, print the output to the terminal
    if echo "$OUTPUT" | grep -q "Error: failed to finish work on issue err: parent issue is not set for"; then
        echo "Found 'success' in the output, exiting..."
        break
    fi
done
