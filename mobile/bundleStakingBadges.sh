#!/bin/bash

# Define the folder containing your images (use Unix-style paths)
IMAGE_FOLDER="../assets/img/stats/staking"
OUTPUT_FILE="bundledStakingIcons.go"

# Ensure fyne is available
if ! command -v fyne &> /dev/null
then
    echo "fyne could not be found. Make sure fyne is installed and in your PATH."
    exit
fi

# Initialize the output file
echo "" > $OUTPUT_FILE

# Loop over each file in the folder
for FILE in "$IMAGE_FOLDER"/*; do
    # Generate the fyne bundle command and append to output file
    fyne bundle -append -o "$OUTPUT_FILE" "$FILE"
done