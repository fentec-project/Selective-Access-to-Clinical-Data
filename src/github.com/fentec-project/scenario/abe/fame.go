/*
 * Copyright (c) 2018 XLAB d.o.o
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

package abe

import (
	"math/big"

	"fmt"
	"strconv"

	"github.com/fentec-project/bn256"
	"github.com/fentec-project/gofe/data"
	"github.com/fentec-project/gofe/sample"
	"github.com/fentec-project/scenario/serialization"
)

// This is a ciphertext policy (CP) attribute based (ABE) scheme based
// on Shashank Agrawal, Melissa Chase:
// "FAME: Fast Attribute-based Message Encryption"
//
// This scheme enables encryption based on a boolean expression
// determining which attributes are needed for an entity to be able
// to decrypt. Moreover, secret keys are generated, where each key
// is connected to some attribute, such that only a set of keys whose
// attributes are sufficient can decrypt the massage.
// This scheme is a PUBLIC-KEY scheme - no master secret key is needed
// to encrypt the messages.
//

// FAME represents a FAME scheme.
type FAME struct {
	P *big.Int // order of the elliptic curve
}

// NewFAME configures a new instance of the scheme.
func NewFAME() *FAME {
	return &FAME{P: bn256.Order}
}

// FAMESecKey represents a master secret key of a FAME scheme.
type FAMESecKey struct {
	PartInt [4]*big.Int
	PartG1  [3]*bn256.G1
}

// FAMEPubKey represents a public key of a FAME scheme.
type FAMEPubKey struct {
	PartG2 [2]*bn256.G2
	PartGT [2]*bn256.GT
}

// GenerateMasterKeys generates a new set of public keys, needed
// for encrypting data, and master secret keys needed for generating
// keys for decrypting.
func GenerateMasterKeys(path string) error {
	// create a new FAME struct with the universe of attributes
	// denoted by integer
	a := NewFAME()
	serialization.WriteGob(path+"fame.gob", a)

	sampler := sample.NewUniformRange(big.NewInt(1), a.P)
	val, err := data.NewRandomVector(7, sampler)
	if err != nil {
		return err
	}

	partInt := [4]*big.Int{val[0], val[1], val[2], val[3]}
	partG1 := [3]*bn256.G1{new(bn256.G1).ScalarBaseMult(val[4]),
		new(bn256.G1).ScalarBaseMult(val[5]),
		new(bn256.G1).ScalarBaseMult(val[6])}
	partG2 := [2]*bn256.G2{new(bn256.G2).ScalarBaseMult(val[0]),
		new(bn256.G2).ScalarBaseMult(val[1])}
	tmp1 := new(big.Int).Mod(new(big.Int).Add(new(big.Int).Mul(val[0], val[4]), val[6]), a.P)
	tmp2 := new(big.Int).Mod(new(big.Int).Add(new(big.Int).Mul(val[1], val[5]), val[6]), a.P)
	partGT := [2]*bn256.GT{new(bn256.GT).ScalarBaseMult(tmp1),
		new(bn256.GT).ScalarBaseMult(tmp2)}

	pk := FAMEPubKey{PartG2: partG2, PartGT: partGT}
	sk := FAMESecKey{PartInt: partInt, PartG1: partG1}
	serialization.WriteGob(path+"pubParams.gob", pk)
	serialization.WriteGob(path+"masterKey.gob", sk)

	return nil
}

// FAMECipher represents a ciphertext of a FAME scheme.
type FAMECipher struct {
	Ct0     [3]*bn256.G2
	Ct      [][3]*bn256.G1
	CtPrime *bn256.GT
	Msp     *MSP
}

// Encrypt takes as an input a message msg represented as an element of an elliptic
// curve, a MSP struct representing the decryption policy, and a public key pk. It
// returns an encryption of the message. In case of a failed procedure an error
// is returned. Note that safety of the encryption is only proved if the mapping
// msp.RowToAttrib from the rows of msp.Mat to attributes is injective.
func Encrypt(path string, msg string, msp *MSP) error {
	msgInGt, err := bn256.MapStringToGT(msg)
	if err != nil {
		return err
	}

	if len(msp.Mat) == 0 || len(msp.Mat[0]) == 0 {
		return fmt.Errorf("empty msp matrix")
	}

	attrib := make(map[int]bool)
	for _, i := range msp.RowToAttrib {
		if attrib[i] {
			return fmt.Errorf("some attributes correspond to" +
				"multiple rows of the MSP struct, the scheme is not secure")
		}
		attrib[i] = true

	}

	a := new(FAME)
	serialization.ReadGob(path+"fame.gob", a)
	pk := new(FAMEPubKey)
	serialization.ReadGob(path+"pubParams.gob", pk)

	sampler := sample.NewUniform(a.P)
	s, err := data.NewRandomVector(2, sampler)
	if err != nil {
		return err
	}
	ct0 := [3]*bn256.G2{new(bn256.G2).ScalarMult(pk.PartG2[0], s[0]),
		new(bn256.G2).ScalarMult(pk.PartG2[1], s[1]),
		new(bn256.G2).ScalarBaseMult(new(big.Int).Add(s[0], s[1]))}

	ct := make([][3]*bn256.G1, len(msp.Mat))
	for i := 0; i < len(msp.Mat); i++ {
		for l := 0; l < 3; l++ {
			hs1, err := bn256.HashG1(strconv.Itoa(msp.RowToAttrib[i]) + " " + strconv.Itoa(l) + " 0")
			if err != nil {
				return err
			}
			hs1.ScalarMult(hs1, s[0])

			hs2, err := bn256.HashG1(strconv.Itoa(msp.RowToAttrib[i]) + " " + strconv.Itoa(l) + " 1")
			if err != nil {
				return err
			}
			hs2.ScalarMult(hs2, s[1])

			ct[i][l] = new(bn256.G1).Add(hs1, hs2)
			for j := 0; j < len(msp.Mat[0]); j++ {
				hs1, err = bn256.HashG1("0 " + strconv.Itoa(j) + " " + strconv.Itoa(l) + " 0")
				if err != nil {
					return err
				}
				hs1.ScalarMult(hs1, s[0])

				hs2, err = bn256.HashG1("0 " + strconv.Itoa(j) + " " + strconv.Itoa(l) + " 1")
				if err != nil {
					return err
				}
				hs2.ScalarMult(hs2, s[1])

				hsToM := new(bn256.G1).Add(hs1, hs2)
				pow := new(big.Int).Set(msp.Mat[i][j])
				if pow.Sign() == -1 {
					pow.Neg(pow)
					hsToM.ScalarMult(hsToM, pow)
					hsToM.Neg(hsToM)
				} else {
					hsToM.ScalarMult(hsToM, pow)
				}
				ct[i][l].Add(ct[i][l], hsToM)
			}
		}
	}

	ctPrime := new(bn256.GT).ScalarMult(pk.PartGT[0], s[0])
	ctPrime.Add(ctPrime, new(bn256.GT).ScalarMult(pk.PartGT[1], s[1]))
	ctPrime.Add(ctPrime, msgInGt)

	cipher := FAMECipher{Ct0: ct0, Ct: ct, CtPrime: ctPrime, Msp: msp}
	serialization.WriteGob(path+"aux_encrypt.gob", cipher)

	return nil
}

// FAMEAttribKeys represents keys corresponding to attributes possessed by
// an entity and used for decrypting in a FAME scheme.
type FAMEAttribKeys struct {
	K0        [3]*bn256.G2
	K         [][3]*bn256.G1
	KPrime    [3]*bn256.G1
	AttribToI map[int]int
}

// GenerateAttribKeys given a set of attributes gamma and the master secret key
// generates keys that can be used for the decryption of any ciphertext encoded
// with a policy for which attributes gamma are sufficient.
func GenerateAttribKeys(path string, keyPath string, gamma []int) error {
	a := new(FAME)
	serialization.ReadGob(path+"fame.gob", a)

	sampler := sample.NewUniform(a.P)
	r, err := data.NewRandomVector(2, sampler)
	if err != nil {
		return err
	}
	sigma, err := data.NewRandomVector(len(gamma), sampler)
	if err != nil {
		return err
	}

	sk := new(FAMESecKey)
	serialization.ReadGob(path+"masterKey.gob", sk)

	pow0 := new(big.Int).Mul(sk.PartInt[2], r[0])
	pow0.Mod(pow0, a.P)
	pow1 := new(big.Int).Mul(sk.PartInt[3], r[1])
	pow1.Mod(pow1, a.P)
	pow2 := new(big.Int).Add(r[0], r[1])
	pow2.Mod(pow2, a.P)

	k0 := [3]*bn256.G2{new(bn256.G2).ScalarBaseMult(pow0),
		new(bn256.G2).ScalarBaseMult(pow1),
		new(bn256.G2).ScalarBaseMult(pow2)}

	a0Inv := new(big.Int).ModInverse(sk.PartInt[0], a.P)
	a1Inv := new(big.Int).ModInverse(sk.PartInt[1], a.P)
	aInv := [2]*big.Int{a0Inv, a1Inv}

	k := make([][3]*bn256.G1, len(gamma))
	attribToI := make(map[int]int)
	for i, y := range gamma {
		k[i] = [3]*bn256.G1{new(bn256.G1), new(bn256.G1), new(bn256.G1)}
		gSigma := new(bn256.G1).ScalarBaseMult(sigma[i])
		for t := 0; t < 2; t++ {
			hs0, err := bn256.HashG1(strconv.Itoa(y) + " 0 " + strconv.Itoa(t))
			if err != nil {
				return err
			}
			hs0.ScalarMult(hs0, pow0)
			hs1, err := bn256.HashG1(strconv.Itoa(y) + " 1 " + strconv.Itoa(t))
			if err != nil {
				return err
			}
			hs1.ScalarMult(hs1, pow1)
			hs2, err := bn256.HashG1(strconv.Itoa(y) + " 2 " + strconv.Itoa(t))
			if err != nil {
				return err
			}
			hs2.ScalarMult(hs2, pow2)

			k[i][t].Add(hs0, hs1)
			k[i][t].Add(k[i][t], hs2)
			k[i][t].Add(k[i][t], gSigma)
			k[i][t].ScalarMult(k[i][t], aInv[t])
		}

		k[i][2].ScalarBaseMult(sigma[i])
		k[i][2].Neg(k[i][2])

		attribToI[y] = i
	}

	sigmaPrime, err := sampler.Sample()
	if err != nil {
		return err
	}
	gSigmaPrime := new(bn256.G1).ScalarBaseMult(sigmaPrime)

	kPrime := [3]*bn256.G1{new(bn256.G1), new(bn256.G1), new(bn256.G1)}
	for t := 0; t < 2; t++ {
		hs0, err := bn256.HashG1("0 0 0 " + strconv.Itoa(t))
		if err != nil {
			return err
		}
		hs0.ScalarMult(hs0, pow0)
		hs1, err := bn256.HashG1("0 0 1 " + strconv.Itoa(t))
		if err != nil {
			return err
		}
		hs1.ScalarMult(hs1, pow1)
		hs2, err := bn256.HashG1("0 0 2 " + strconv.Itoa(t))
		if err != nil {
			return err
		}
		hs2.ScalarMult(hs2, pow2)

		kPrime[t].Add(hs0, hs1)
		kPrime[t].Add(kPrime[t], hs2)
		kPrime[t].Add(kPrime[t], gSigmaPrime)
		kPrime[t].ScalarMult(kPrime[t], aInv[t])
		kPrime[t].Add(kPrime[t], sk.PartG1[t])

	}
	kPrime[2].ScalarBaseMult(sigmaPrime)
	kPrime[2].Neg(kPrime[2])
	kPrime[2].Add(kPrime[2], sk.PartG1[2])

	key := FAMEAttribKeys{K0: k0, K: k, KPrime: kPrime, AttribToI: attribToI}
	serialization.WriteGob(path+keyPath, key)

	return nil
}

// Decrypt takes as an input a cipher and an FAMEAttribKeys and tries to decrypt
// the cipher. This is possible only if the set of possessed attributes (and
// corresponding keys FAMEAttribKeys) suffices the encryption policy of the
// cipher. If this is not possible, an error is returned.
func Decrypt(path string, keyPath string) error {
	a := new(FAME)
	serialization.ReadGob(path+"fame.gob", a)
	cipher := new(FAMECipher)
	serialization.ReadGob(path+"aux_encrypt.gob", cipher)
	key := new(FAMEAttribKeys)
	serialization.ReadGob(path+keyPath, key)

	// find out which attributes are owned
	attribMap := make(map[int]bool)
	for k := range key.AttribToI {
		attribMap[k] = true
	}

	countAttrib := 0
	for i := 0; i < len(cipher.Msp.Mat); i++ {
		if attribMap[cipher.Msp.RowToAttrib[i]] {
			countAttrib++
		}
	}

	// create a matrix of needed keys
	preMatForKey := make([]data.Vector, countAttrib)
	ctForKey := make([][3]*bn256.G1, countAttrib)
	rowToAttrib := make([]int, countAttrib)
	countAttrib = 0
	for i := 0; i < len(cipher.Msp.Mat); i++ {
		if attribMap[cipher.Msp.RowToAttrib[i]] {
			preMatForKey[countAttrib] = cipher.Msp.Mat[i]
			ctForKey[countAttrib] = cipher.Ct[i]
			rowToAttrib[countAttrib] = cipher.Msp.RowToAttrib[i]
			countAttrib++
		}
	}

	matForKey, err := data.NewMatrix(preMatForKey)
	if err != nil {
		return fmt.Errorf("the provided cipher is faulty")
	}

	// get a combination alpha of keys needed to decrypt
	oneVec := data.NewConstantVector(len(matForKey[0]), big.NewInt(0))
	oneVec[0].SetInt64(1)
	alpha, err := gaussianElimination(matForKey.Transpose(), oneVec, a.P)
	if err != nil {
		return fmt.Errorf("provided key is not sufficient for decryption")
	}

	msgInGt := new(bn256.GT).Set(cipher.CtPrime)

	ctProd := new([3]*bn256.G1)
	keyProd := new([3]*bn256.G1)
	for j := 0; j < 3; j++ {
		ctProd[j] = new(bn256.G1).ScalarBaseMult(big.NewInt(0))
		keyProd[j] = new(bn256.G1).ScalarBaseMult(big.NewInt(0))
		for i, e := range rowToAttrib {
			ctProd[j].Add(ctProd[j], new(bn256.G1).ScalarMult(ctForKey[i][j], alpha[i]))
			keyProd[j].Add(keyProd[j], new(bn256.G1).ScalarMult(key.K[key.AttribToI[e]][j], alpha[i]))
		}
		keyProd[j].Add(keyProd[j], key.KPrime[j])
		ctPairing := bn256.Pair(ctProd[j], key.K0[j])
		keyPairing := bn256.Pair(keyProd[j], cipher.Ct0[j])
		keyPairing.Neg(keyPairing)
		msgInGt.Add(msgInGt, ctPairing)
		msgInGt.Add(msgInGt, keyPairing)
	}

	decipher := bn256.MapGTToString(msgInGt)
	serialization.WriteGob(path+"aux_decrypt.gob", decipher)

	return nil
}
