package main

import (
	"fmt"

	"assign2/internal/constants"
	"assign2/internal/messages"
)

func validateJob(job *messages.JobMsg) error {
	// if job.PasswordLen != constants.PasswordLen {
	// 	return fmt.Errorf("invalid password length")
	// }
	if job.Charset != constants.LegalCharset79 {
		return fmt.Errorf("invalid charset")
	}
	switch job.Alg {
	case "yescrypt", "bcrypt", "sha256", "sha512", "md5":
	default:
		return fmt.Errorf("unsupported hash algorithm")
	}
	if job.FullHash == "" {
		return fmt.Errorf("empty hash field")
	}
	return nil
}
