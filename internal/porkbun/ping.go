// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package porkbun

import "context"

func (c *Client) Ping(
	ctx context.Context,
) (
	string,
	error,
) {
	path := "ping"
	req := &Credentials{}
	var res struct {
		Status
		YourIP string `json:"yourIp"`
	}
	err := c.post(ctx, path, req, &res)
	return res.YourIP, err
}
