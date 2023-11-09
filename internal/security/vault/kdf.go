/*
MIT License

Copyright (c) 2018 Alex Edwards

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package vault

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"

	"golang.org/x/crypto/argon2"
)

type ArgonParameters struct {
	Aargon2Version uint32
	Variant        string
	Memory         uint32
	Iterations     uint32
	Parallelism    uint8
	SaltLength     uint32
	KeyLength      uint32
}

var DefaultParameters = &ArgonParameters{
	Aargon2Version: argon2.Version,
	Variant:        "argon2id",
	Memory:         64 * 1024,
	Iterations:     3,
	Parallelism:    2,
	SaltLength:     16,
	KeyLength:      32,
}

func generateFromPassword(password string, p *ArgonParameters) (hash, salt []byte, err error) {
	salt, err = generateRandomBytes(p.SaltLength)
	if err != nil {
		return nil, nil, err
	}

	hash = argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	return hash, salt, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func comparePasswordAndHash(password string, salt, hash []byte, p *ArgonParameters) (bool, []byte, error) {
	derivedKey := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	derivedKeyHash := sha256Hash(derivedKey)

	if subtle.ConstantTimeCompare(hash, derivedKeyHash) == 1 {
		return true, derivedKey, nil
	}
	return false, nil, nil
}

func sha256Hash(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}
