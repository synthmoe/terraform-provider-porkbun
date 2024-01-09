// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package porkbun

import (
	"context"
	"encoding/json"
	"fmt"
)

type Domain struct {
	Domain       string
	Status       string
	TLD          string
	CreateDate   string
	ExpireDate   string
	SecurityLock bool
	WhoisPrivacy bool
	AutoRenew    bool
	NotLocal     bool
}

type domain struct {
	Domain       string      `json:"domain"`
	Status       string      `json:"status"`
	TLD          string      `json:"TLD"`
	CreateDate   string      `json:"createDate"`
	ExpireDate   string      `json:"expireDate"`
	SecurityLock json.Number `json:"securityLock"`
	WhoisPrivacy json.Number `json:"whoisPrivacy"`
	AutoRenew    json.Number `json:"autoRenew"`
	NotLocal     json.Number `json:"notLocal"`
}

func (d *domain) convert() (*Domain, error) {
	domain := &Domain{
		Domain:     d.Domain,
		Status:     d.Status,
		TLD:        d.TLD,
		CreateDate: d.CreateDate,
		ExpireDate: d.ExpireDate,
	}
	switch d.SecurityLock.String() {
	case "1":
		domain.SecurityLock = true
	case "0":
		domain.SecurityLock = false
	default:
		return nil, fmt.Errorf(
			"Expected security lock of '1' or '0', got '%s'.",
			d.SecurityLock.String(),
		)
	}
	switch d.NotLocal.String() {
	case "1":
		domain.WhoisPrivacy = true
	case "0":
		domain.WhoisPrivacy = false
	default:
		return nil, fmt.Errorf(
			"Expected whois privacy of '1' or '0', got '%s'.",
			d.WhoisPrivacy.String(),
		)
	}
	switch d.NotLocal.String() {
	case "1":
		domain.AutoRenew = true
	case "0":
		domain.AutoRenew = false
	default:
		return nil, fmt.Errorf(
			"Expected auto renew of '1' or '0', got '%s'.",
			d.AutoRenew.String(),
		)
	}
	switch d.NotLocal.String() {
	case "1":
		domain.NotLocal = true
	case "0":
		domain.NotLocal = false
	default:
		return nil, fmt.Errorf(
			"Expected not local of '1' or '0', got '%s'.",
			d.NotLocal.String(),
		)
	}
	return domain, nil
}

func (c *Client) DomainList(
	ctx context.Context,
	start int64,
) (
	[]*Domain,
	[]error,
) {
	path := "domain/listAll"
	req := &struct {
		Credentials
		Start int64 `json:"start"`
	}{Start: start}
	var res struct {
		Status
		Domains []domain `json:"domains"`
	}
	err := c.post(ctx, path, req, &res)
	if err != nil {
		return nil, []error{err}
	}
	domains := []*Domain{}
	errs := []error{}
	for _, domain := range res.Domains {
		d, err := domain.convert()
		if err != nil {
			errs = append(errs, err)
		}
		domains = append(domains, d)
	}
	return domains, errs
}
