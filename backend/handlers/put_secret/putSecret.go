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
	"fmt"
	"seif/crypton"
	"seif/db_ops"
	"seif/params"
	"seif/utils"

	"github.com/gofiber/fiber/v2"
)

type request struct {
	Secret string `json:"secret"`
	Expiry int    `json:"expiry"`
}

type response struct {
	Id  string `json:"id"`
	Key string `json:"key"`
}

const SQL = `
	INSERT INTO SECRETS (ID, SECRET, EXPIRY, TS)
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`

func PutSecret(c *fiber.Ctx) error {
	req := new(request)
	if err := c.BodyParser(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, utils.FHE004, "body", &err)
	}

	if len(req.Secret) > params.MaxBytes {
		return utils.SendError(c, fiber.StatusBadRequest, utils.FHE005, "", nil)
	}

	if req.Expiry < 1 || req.Expiry > params.MaxDays {
		return utils.SendError(c, fiber.StatusBadRequest, utils.FHE006, fmt.Sprint(params.MaxDays), nil)
	}

	id, key, crypto, err := crypton.Encode(req.Secret)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, utils.FHE007, "", &err)
	}

	ret := response{Id: crypton.Bs2str(id), Key: crypton.Bs2str(key)}

	defer func() { go db_ops.Backup() }()
	params.Lock.Lock()
	defer params.Lock.Unlock()

	_, err = params.Db.Exec(SQL, ret.Id, crypto, req.Expiry)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, utils.FHE002, "secrets", &err)
	}

	c.JSON(ret)
	return c.SendStatus(fiber.StatusOK)
}
