#!/bin/bash

# This script is used to run the catgo program.
catgo() {
  local source_dir="$1"
  local output_file="$2"
  local file_extension="$3"
  find "$source_dir" -type f -name "*.$file_extension" -print0 |
    while IFS= read -r -d $'\0' file; do
      cat "$file" >> "$output_file"
      echo "/********************************END_OF_FILE************************************/" >> "$output_file"
    done
}

# Usage
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <directory>"
  exit 1
fi

source_dir="$1"
output_file="output.go"
file_extension="go"

# Call the function to concatenate files
catgo "$source_dir" "$output_file" "$file_extension"

echo "Concatenated files are saved in $output_file"
