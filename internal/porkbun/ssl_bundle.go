// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package porkbun

import "context"

type SSLBundle struct {
	IntermediateCertificate string `json:"intermediatecertificate"`
	CertificateChain        string `json:"certificatechain"`
	PublicKey               string `json:"publickey"`
	PrivateKey              string `json:"privatekey"`
}

func (c *Client) SSLBundle(
	ctx context.Context,
	domain string,
) (
	*SSLBundle,
	error,
) {
	path := "ssl/retrieve/" + domain
	req := &Credentials{}
	var res struct {
		Status
		SSLBundle
	}
	err := c.post(ctx, path, req, &res)
	return &res.SSLBundle, err
}
