#!/bin/bash

# Define the folder containing your images (use Unix-style paths)
IMAGE_FOLDER="../assets/img/placeholder.png"
OUTPUT_FILE="bundledPlaceHolderIcon.go"

# Ensure fyne is available
if ! command -v fyne &> /dev/null
then
    echo "fyne could not be found. Make sure fyne is installed and in your PATH."
    exit 1
fi

# Bundle the image file
fyne bundle -o "$OUTPUT_FILE" "$IMAGE_FOLDER"
