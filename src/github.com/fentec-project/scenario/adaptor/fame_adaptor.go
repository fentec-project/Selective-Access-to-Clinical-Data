package adaptor

import ()

/*
 * Copyright (c) 2018 ATOS
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import (
	"fmt"
	"github.com/fentec-project/gofe/abe"
	"github.com/fentec-project/scenario/serialization"
)

// This is an adaptor for the fame ABE scheme on github.com/fentec-project/gofe/abe
// in order to save and load keys from files

func GenerateMasterKeys(path string) error {
	// create a new FAME struct with the universe of attributes
	// denoted by integer
	a := abe.NewFAME()
	serialization.WriteGob(path+"fame.gob", a)

	pubKey, secKey, err := a.GenerateMasterKeys()
	if err != nil {
		fmt.Errorf("Failed to generate master key: %v", err)
	}
	serialization.WriteGob(path+"pubParams.gob", pubKey)
	serialization.WriteGob(path+"masterKey.gob", secKey)

	return nil
}

func Encrypt(path string, msg string, msp *abe.MSP) error {

	a := new(abe.FAME)
	serialization.ReadGob(path+"fame.gob", a)
	pk := new(abe.FAMEPubKey)
	serialization.ReadGob(path+"pubParams.gob", pk)

	cipher, err := a.Encrypt(msg, msp, pk)
	if err != nil {
		fmt.Errorf("Failed to encrypt: %v", err)
	}

	serialization.WriteGob(path+"aux_encrypt.gob", *cipher)

	return nil
}

func GenerateAttribKeys(path string, keyPath string, gamma []int) error {
	a := new(abe.FAME)
	serialization.ReadGob(path+"fame.gob", a)

	sk := new(abe.FAMESecKey)
	serialization.ReadGob(path+"masterKey.gob", sk)

	keys, err := a.GenerateAttribKeys(gamma, sk)
	if err != nil {
		fmt.Errorf("Failed to generate keys: %v", err)
	}

	serialization.WriteGob(path+keyPath, keys)

	return nil
}

// Decrypt takes as an input a cipher and an FAMEAttribKeys and tries to decrypt
// the cipher. This is possible only if the set of possessed attributes (and
// corresponding keys FAMEAttribKeys) suffices the encryption policy of the
// cipher. If this is not possible, an error is returned.
func Decrypt(path string, keyPath string) error {
	a := new(abe.FAME)
	serialization.ReadGob(path+"fame.gob", a)
	cipher := new(abe.FAMECipher)
	serialization.ReadGob(path+"aux_encrypt.gob", cipher)
	key := new(abe.FAMEAttribKeys)
	serialization.ReadGob(path+keyPath, key)
	pk := new(abe.FAMEPubKey)
	serialization.ReadGob(path+"pubParams.gob", pk)

	msgCheck, err := a.Decrypt(cipher, key, pk)
	if err != nil {
		fmt.Errorf("Failed to decrypt: %v", err)
		return err
	}

	serialization.WriteGob(path+"aux_decrypt.gob", msgCheck)

	return nil
}
