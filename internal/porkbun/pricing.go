// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package porkbun

import "context"

type Pricing struct {
	Registration string `json:"registration"`
	Renewal      string `json:"renewal"`
	Transfer     string `json:"transfer"`
	SpecialType  string `json:"specialType"`
}

func (c *Client) Pricing(
	ctx context.Context,
) (
	map[string]Pricing,
	error,
) {
	path := "pricing/get"
	var res struct {
		Status
		pricing map[string]Pricing `json:"pricing"`
	}
	err := c.get(ctx, path, &res)
	return res.pricing, err
}
