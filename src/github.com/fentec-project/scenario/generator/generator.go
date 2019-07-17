package main

import (
	"fmt"
	"github.com/fentec-project/scenario/abe"
)

func main() {
	debug := true

	path := "key-material/"

	// generate a public key and a secret key for the scheme
	abe.GenerateMasterKeys(path)

	/* Decryptor 1 */
	if debug {
		fmt.Println("Key for decryptor 1: {0, 2, 3}")
	}
	// define a set of attributes (a subset of the universe of attributes)
	// that an entity possesses
	gamma := []int{0, 2, 3}
	// generate keys for decryption for an entity with
	// attributes gamma
	abe.GenerateAttribKeys(path, "keyDecryptor1.gob", gamma)

	/* Decryptor 2 */
	if debug {
		fmt.Println("Key for decryptor 2: {0, 4}")
	}
	// define a set of attributes (a subset of the universe of attributes)
	// that an entity possesses
	gamma = []int{0, 4}
	// generate keys for decryption for an entity with
	// attributes gamma
	abe.GenerateAttribKeys(path, "keyDecryptor2.gob", gamma)

	/* Decryptor 3 */
	if debug {
		fmt.Println("Key for decryptor 3: {0, 2, 5}")
	}
	// define a set of attributes (a subset of the universe of attributes)
	// that an entity possesses
	gamma = []int{0, 2, 5}
	// generate keys for decryption for an entity with
	// attributes gamma
	abe.GenerateAttribKeys(path, "keyDecryptor3.gob", gamma)

	fmt.Println("\n---------- GENERATOR finished ----------\n")
}
