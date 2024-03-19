/*
 * Copyright (C) 2024- Germano Rizzo
 *
 * This file is part of Seif.
 *
 * Seif is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Seif is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with Seif.  If not, see <http://www.gnu.org/licenses/>.
 */
package crypton

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
)

const IV_LEN = 48 >> 3
const KEY_LEN = 96 >> 3

const IV_LEN_COMPLETE = 12 // FIXME must be a real constant
const KEY_LEN_COMPLETE = aes.BlockSize

func genRandomBytes(length int) ([]byte, error) {
	ret := make([]byte, length)

	if _, err := rand.Read(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func lengthen(bs []byte, length int) []byte {
	ret := make([]byte, length)
	copy(ret, bs)
	return ret
}

func Bs2str(bs []byte) string {
	return base64.URLEncoding.EncodeToString(bs)
}

func Str2bs(str string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(str)
}

func Encode(message string) (id []byte, key []byte, crypto []byte, err error) {
	key, err = genRandomBytes(KEY_LEN)
	if err != nil {
		return nil, nil, nil, err
	}
	key2 := lengthen(key, KEY_LEN_COMPLETE)

	aesBlock, err := aes.NewCipher(key2)
	if err != nil {
		return nil, nil, nil, err
	}

	aesgcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, nil, nil, err
	}

	id, err = genRandomBytes(IV_LEN)
	if err != nil {
		return nil, nil, nil, err
	}
	id2 := lengthen(id, IV_LEN_COMPLETE)

	crypto = aesgcm.Seal(nil, id2, []byte(message), nil)

	return
}

func Decode(id []byte, key []byte, crypto []byte) (message string, err error) {
	aesBlock, err := aes.NewCipher(lengthen(key, KEY_LEN_COMPLETE))
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return "", err
	}

	plain, err := aesgcm.Open(nil, lengthen(id, IV_LEN_COMPLETE), crypto, nil)
	if err != nil {
		return "", err
	}

	message = string(plain)

	return
}
