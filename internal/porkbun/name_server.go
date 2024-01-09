// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package porkbun

import "context"

func (c *Client) NameServers(
	ctx context.Context,
	domain string,
) (
	[]string,
	error,
) {
	path := "domain/getNs/" + domain
	req := &Credentials{}
	var res struct {
		Status
		NS []string `json:"ns"`
	}
	err := c.post(ctx, path, req, &res)
	return res.NS, err
}

func (c *Client) UpdateNameServers(
	ctx context.Context,
	domain string,
	ns []string,
) error {
	path := "domain/updateNs/" + domain
	req := &struct {
		Credentials
		NS []string `json:"ns"`
	}{NS: ns}
	var res Status
	err := c.post(ctx, path, req, &res)
	return err
}
