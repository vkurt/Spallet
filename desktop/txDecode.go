package main

import (
	"encoding/hex"
	"fmt"
	"unicode/utf8"
)

// hexToASCII converts hexadecimal string to ASCII, adding a space for unreadable characters
func hexToASCII(hexStr string) string {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		fmt.Println("Error decoding hex string:", err)
		return ""
	}

	// Convert hex bytes to ASCII with spaces for unreadable characters
	var result string
	for _, b := range bytes {
		if utf8.Valid([]byte{b}) && (b >= 32 && b <= 126) { // printable ASCII range
			result += string(b)
		} else {
			result += " " // Add space for unreadable characters
		}
	}
	return result
}

// func decodeData(data string) {
// 	// Your hex data
// 	// Convert hex data to readable text
// 	readableText := hexToASCII(data)
// 	fmt.Println(readableText)
// }
