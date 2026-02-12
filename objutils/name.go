package objutils

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"unicode"

	"github.com/mandelsoft/goutils/general"
	"k8s.io/apimachinery/pkg/util/validation"
)

const MAX_NAMELEN = 253
const MAX_NAMESPACELEN = 63

func checkName(name string) {
	errs := validation.IsDNS1123Subdomain(name)
	if len(errs) > 0 {
		fmt.Printf("Ungültiger Name: %v\n", errs)
	} else {
		fmt.Println("Name ist valide!")
	}
}

func GenerateUniqueName(prefix, cluster, namespace, name string, length ...int) string {
	// 0. Cleanup name
	n := ""
	for _, r := range cluster {
		if unicode.IsDigit(r) || (unicode.IsLetter(r) && r < 256) {
			n += string(unicode.ToLower(r))
		} else {
			n += "--"
		}
	}

	// 1. Create a hash of the unique parts
	var fullInput string
	if namespace == "" {
		fullInput = name

	} else {
		fullInput = fmt.Sprintf("%s-%s", namespace, name)
	}
	if n != "" {
		fullInput = fmt.Sprintf("%s-%s", fullInput, n)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(fullInput)))[:8] // Take first 8 chars

	// 2. Sanitize and truncate the name
	// We want: prefix (variable) + name (truncated) + "-" + hash (8 chars)
	// Total must be <= 63 for safe usage across all resource types
	limit := general.OptionalDefaulted(MAX_NAMELEN, length...)
	maxNameLen := limit - len(prefix) - len(hash) - 2 // -2 for the separators

	shortName := fullInput
	if len(shortName) > maxNameLen {
		shortName = fullInput[:maxNameLen]
	}

	// Clean up trailing dashes if truncation cut in the middle
	shortName = strings.TrimSuffix(shortName, "-")

	return fmt.Sprintf("%s-%s-%s", prefix, shortName, hash)
}
