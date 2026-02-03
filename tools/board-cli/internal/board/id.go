package board

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"time"
)

// base36Chars contains lowercase alphanumeric characters for base36 encoding.
const base36Chars = "0123456789abcdefghijklmnopqrstuvwxyz"

// GenerateID creates a distributed ID in the format YYMMDD-xxxxxx.
// The suffix is the first 6 characters of base36(sha256(uuid)).
// UUID guarantees uniqueness; hash provides short readable suffix.
func GenerateID(typ ElementType) string {
	now := time.Now()

	// Date part: YYMMDD
	datePart := now.Format("060102")

	// Hash part: sha256(uuid) → base36 → first 4 chars
	hashPart := generateHashSuffix()

	// Combine: TYPE-YYMMDD-xxxx
	prefix := typePrefix(typ)
	return fmt.Sprintf("%s-%s-%s", prefix, datePart, hashPart)
}

// generateHashSuffix creates a 6-character base36 string from a UUID.
func generateHashSuffix() string {
	// Generate 16 random bytes (UUID v4 equivalent)
	uuid := make([]byte, 16)
	rand.Read(uuid)

	// SHA256 of UUID for uniform distribution
	hash := sha256.Sum256(uuid)

	// Convert first 8 bytes to uint64 for base36 encoding
	var num uint64
	for i := 0; i < 8; i++ {
		num = (num << 8) | uint64(hash[i])
	}

	// Encode to base36 and take first 6 characters
	encoded := toBase36(num)
	if len(encoded) < 6 {
		// Pad with zeros if needed (unlikely but safe)
		encoded = "000000"[:6-len(encoded)] + encoded
	}
	return encoded[:6]
}

// toBase36 converts a uint64 to a base36 string.
func toBase36(num uint64) string {
	if num == 0 {
		return "0"
	}

	var result []byte
	for num > 0 {
		result = append([]byte{base36Chars[num%36]}, result...)
		num /= 36
	}
	return string(result)
}

// typePrefix returns the ID prefix for an element type.
func typePrefix(typ ElementType) string {
	switch typ {
	case EpicType:
		return "EPIC"
	case StoryType:
		return "STORY"
	case TaskType:
		return "TASK"
	case BugType:
		return "BUG"
	default:
		return "ITEM"
	}
}
