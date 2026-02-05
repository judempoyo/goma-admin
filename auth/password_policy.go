package auth

import (
	"fmt"
	"strings"
	"unicode"
)

type PasswordPolicy struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

type PasswordPolicyError struct {
	Reason string
}

func (e PasswordPolicyError) Error() string {
	return e.Reason
}

func (p PasswordPolicy) Validate(email, password string) error {
	if p.MinLength == 0 {
		p.MinLength = 12
	}
	if p.MaxLength == 0 {
		p.MaxLength = 128
	}
	if len(password) < p.MinLength {
		return PasswordPolicyError{Reason: fmt.Sprintf("password must be at least %d characters", p.MinLength)}
	}
	if len(password) > p.MaxLength {
		return PasswordPolicyError{Reason: fmt.Sprintf("password must be at most %d characters", p.MaxLength)}
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if p.RequireUpper && !hasUpper {
		return PasswordPolicyError{Reason: "password must include an uppercase letter"}
	}
	if p.RequireLower && !hasLower {
		return PasswordPolicyError{Reason: "password must include a lowercase letter"}
	}
	if p.RequireNumber && !hasNumber {
		return PasswordPolicyError{Reason: "password must include a number"}
	}
	if p.RequireSpecial && !hasSpecial {
		return PasswordPolicyError{Reason: "password must include a special character"}
	}

	local := strings.ToLower(strings.TrimSpace(email))
	if local != "" {
		parts := strings.Split(local, "@")
		if len(parts) > 0 && len(parts[0]) >= 3 {
			if strings.Contains(strings.ToLower(password), parts[0]) {
				return PasswordPolicyError{Reason: "password must not contain the email local-part"}
			}
		}
	}
	return nil
}
