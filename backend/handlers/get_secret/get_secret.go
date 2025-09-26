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
package get_secret

import (
	"context"
	"seif/crypton"
	"seif/params"
	"seif/utils"

	"github.com/gofiber/fiber/v2"
)

type response struct {
	Secret *string `json:"secret"`
}

const SQL1 = "SELECT SECRET FROM SECRETS WHERE ID = $1"
const SQL2 = "DELETE FROM SECRETS WHERE ID = $1"

func GetSecret(c *fiber.Ctx) error {
	id := c.Query("id", "")
	idBs, err := crypton.Str2bs(id)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, utils.FHE004, "id", &err)
	}

	key := c.Query("key", "")
	keyBs, err := crypton.Str2bs(key)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, utils.FHE004, "key", &err)
	}

	params.Lock.Lock()
	defer params.Lock.Unlock()

	ret := response{}

	tx, err := params.Db.BeginTx(context.Background(), nil)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "FHE007", "", &err)
	}
	defer tx.Rollback()

	rows, err := tx.Query(SQL1, id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, utils.FHE001, "secret", &err)
	}
	defer rows.Close()
	if rows.Next() {
		var secret []byte
		err = rows.Scan(&secret)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, utils.FHE001, "secret", &err)
		}

		plaintxt, err := crypton.Decode(idBs, keyBs, secret)
		if err != nil {
			return utils.SendError(c, fiber.StatusBadRequest, utils.FHE008, "decryption", &err)
		}

		ret.Secret = &plaintxt
	}
	if err = rows.Err(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, utils.FHE003, "secret", &err)
	}

	if ret.Secret != nil {
		_, err := tx.Exec(SQL2, id)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, utils.FHE009, "secret", &err)
		}
	}

	if err := tx.Commit(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, utils.FHE009, "transaction", &err)
	}

	c.JSON(ret)
	return c.SendStatus(fiber.StatusOK)
}
