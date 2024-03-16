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
package put_secret

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"seif/db_ops"
	"seif/handlers/get_secret"
	"seif/params"
	"seif/utils"

	"github.com/gofiber/fiber/v2"
)

type request struct {
	get_secret.Secret
	Expiry int `json:"expiry"`
}

type response struct {
	Id string `json:"id"`
}

const SQL = `
	INSERT INTO SECRETS (ID, IV, SECRET, SHA, EXPIRY, TS)
	VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
	RETURNING ID`

// generateRandomBase64 generates a 42-bit random value encoded in base64.
func generate42bitRandomBase64() (string, error) {
	// 42 bits = 5.25 bytes, but we need a whole number of bytes
	b := make([]byte, 6) // Using 6 bytes (48 bits) to have a whole number greater than 42 bits
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Mask the last 6 bits of the last byte to zero to ensure only 42 bits are random
	b[5] &= 0xC0 // 0xC0 is 11000000 in binary, which sets the last 6 bits to zero

	// Encode to base64
	encoded := base64.URLEncoding.EncodeToString(b)
	// Trim the result to the correct length: 7 characters for 42 bits
	return encoded[:7], nil
}

func PutSecret(c *fiber.Ctx) error {
	req := new(request)
	if err := c.BodyParser(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, utils.FHE004, "body", &err)
	}

	if len(req.Sec) > params.MaxBytes {
		return utils.SendError(c, fiber.StatusBadRequest, utils.FHE005, "", nil)
	}

	if req.Expiry < 1 || req.Expiry > params.MaxDays {
		return utils.SendError(c, fiber.StatusBadRequest, utils.FHE006, fmt.Sprint(params.MaxDays), nil)
	}

	id, err := generate42bitRandomBase64()
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, utils.FHE007, "", &err)
	}

	defer func() { go db_ops.Backup() }()
	params.Lock.Lock()
	defer params.Lock.Unlock()

	_, err = params.Db.Exec(SQL, id, req.IV, req.Sec, req.SHA, req.Expiry)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, utils.FHE002, "secrets", &err)
	}

	c.JSON(response{Id: id})
	return c.SendStatus(fiber.StatusOK)
}
