package kafka

import "crypto/rand"

// generateUUID generates a UUID v4 string using crypto/rand.
// Returns a string in format "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".
func generateUUID() string {
	var uuid [16]byte
	_, _ = rand.Read(uuid[:])

	// Set version to 4 (bits 4-7 of byte 6)
	uuid[6] = (uuid[6] & 0x0f) | 0x40

	// Set variant to RFC 4122 (bits 6-7 of byte 8 set to 10)
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	// Format as xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	const hex = "0123456789abcdef"
	var buf [36]byte
	pos := 0
	for i, b := range uuid {
		if i == 4 || i == 6 || i == 8 || i == 10 {
			buf[pos] = '-'
			pos++
		}
		buf[pos] = hex[b>>4]
		buf[pos+1] = hex[b&0x0f]
		pos += 2
	}
	return string(buf[:])
}
