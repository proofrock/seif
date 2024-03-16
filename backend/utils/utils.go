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
package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

func Abort(msg string, a ...any) {
	fmt.Fprintf(os.Stderr, "FATAL: %s\n", fmt.Sprintf(msg, a...))
	os.Exit(-1)
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

type errorr struct {
	Code   string  `json:"code"`
	Object string  `json:"object"`
	Error  *string `json:"error"`
}

func SendError(c *fiber.Ctx, status int, errCode string, obj string, err *error) error {
	var errString *string
	if err != nil {
		_errString := (*err).Error()
		errString = &_errString
	}
	e := errorr{
		Code:   errCode,
		Object: obj,
		Error:  errString,
	}

	str, _ := json.Marshal(e)
	fmt.Fprintf(os.Stderr, "%s\n", str)
	c.JSON(e)
	return c.SendStatus(status)
}
