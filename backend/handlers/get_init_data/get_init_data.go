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
package get_init_data

import (
	"seif/params"

	"github.com/gofiber/fiber/v2"
)

type response struct {
	Version string `json:"version"`
	MaxDays int    `json:"max_days"`
}

func GetInitData(c *fiber.Ctx) error {
	c.JSON(response{Version: params.VERSION, MaxDays: params.MaxDays})
	return c.SendStatus(fiber.StatusOK)
}
