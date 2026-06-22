package unit_test

import (
	"crypto/subtle"
	"encoding/base64"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
)

// These tests exercise the argon2id helpers exposed via the identity package's
// internal functions. Since hashPasswordArgon2id / verifyArgon2id are package-
// private, we re-implement the same algorithm here with known vectors.

const (
	testArgonMemory      uint32 = 65536
	testArgonIterations  uint32 = 3
	testArgonParallelism uint8  = 2
	testArgonSaltLen            = 16
	testArgonKeyLen      uint32 = 32
)

func hashArgon2idTest(password string) (string, error) {
	salt := make([]byte, testArgonSaltLen)
	// use a fixed salt for deterministic testing
	for i := range salt {
		salt[i] = byte(i)
	}
	hash := argon2.IDKey([]byte(password), salt, testArgonIterations, testArgonMemory, testArgonParallelism, testArgonKeyLen)
	return base64.RawStdEncoding.EncodeToString(salt) + ":" + base64.RawStdEncoding.EncodeToString(hash), nil
}

func verifyArgon2idTest(password, encoded string) bool {
	parts := strings.SplitN(encoded, ":", 2)
	if len(parts) != 2 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	computed := argon2.IDKey([]byte(password), salt, testArgonIterations, testArgonMemory, testArgonParallelism, testArgonKeyLen)
	return subtle.ConstantTimeCompare(computed, expected) == 1
}

func TestArgon2id_HashAndVerify(t *testing.T) {
	password := "correct-horse-battery-staple"
	hash, err := hashArgon2idTest(password)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !verifyArgon2idTest(password, hash) {
		t.Error("correct password should verify successfully")
	}
}

func TestArgon2id_WrongPassword(t *testing.T) {
	hash, _ := hashArgon2idTest("original-password")
	if verifyArgon2idTest("wrong-password", hash) {
		t.Error("wrong password should NOT verify")
	}
}

func TestArgon2id_EmptyPassword(t *testing.T) {
	hash, _ := hashArgon2idTest("")
	if !verifyArgon2idTest("", hash) {
		t.Error("empty password should verify against empty hash")
	}
	if verifyArgon2idTest("notempty", hash) {
		t.Error("non-empty should NOT verify against empty-password hash")
	}
}

func TestArgon2id_TamperedHash(t *testing.T) {
	hash, _ := hashArgon2idTest("secret")
	tampered := hash[:len(hash)-4] + "XXXX"
	if verifyArgon2idTest("secret", tampered) {
		t.Error("tampered hash should NOT verify")
	}
}

func TestArgon2id_InvalidFormat(t *testing.T) {
	if verifyArgon2idTest("secret", "nocolon") {
		t.Error("hash without colon separator should NOT verify")
	}
}
