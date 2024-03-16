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
	"seif/db_ops"
	"seif/params"
	"seif/utils"

	"github.com/gofiber/fiber/v2"
)

type Secret struct {
	IV  string `json:"iv"`
	Sec string `json:"secret"`
	SHA string `json:"sha"`
}

type response struct {
	Secret *Secret `json:"secret"`
}

const SQL = "DELETE FROM SECRETS WHERE ID = $1 RETURNING IV, SECRET, SHA"

func GetSecret(c *fiber.Ctx) error {
	id := c.Query("id", "")

	defer func() { go db_ops.Backup() }()
	params.Lock.Lock()
	defer params.Lock.Unlock()

	ret := response{}

	rows, err := params.Db.Query(SQL, id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, utils.FHE002, "secret", &err)
	}
	defer rows.Close()
	if rows.Next() {
		var secret Secret
		err = rows.Scan(&secret.IV, &secret.Sec, &secret.SHA)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, utils.FHE001, "secret", &err)
		}
		ret.Secret = &secret
	}
	if err = rows.Err(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, utils.FHE003, "secret", &err)
	}

	c.JSON(ret)
	return c.SendStatus(fiber.StatusOK)
}
