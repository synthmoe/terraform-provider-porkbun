// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package porkbun

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type URLForward struct {
	ID          *int64
	Subdomain   string
	Location    string
	Type        string
	IncludePath bool
	Wildcard    bool
}

type urlforward struct {
	ID          json.Number `json:"id,omitempty"`
	Subdomain   string      `json:"subdomain"`
	Location    string      `json:"location"`
	Type        string      `json:"type"`
	IncludePath string      `json:"type"`
	Wildcard    string      `json:"wildcard"`
}

func (f *URLForward) convert() *urlforward {
	id := ""
	if f.ID != nil {
		id = strconv.FormatInt(*f.ID, 10)
	}
	includePath := "no"
	if f.IncludePath {
		includePath = "yes"
	}
	wildcard := "no"
	if f.Wildcard {
		wildcard = "yes"
	}
	return &urlforward{
		ID:          json.Number(id),
		Subdomain:   f.Subdomain,
		Location:    f.Location,
		Type:        f.Type,
		IncludePath: includePath,
		Wildcard:    wildcard,
	}
}

func (f *urlforward) convert() (*URLForward, error) {
	id, err := f.ID.Int64()
	if err != nil {
		return nil, fmt.Errorf(
			"Failed to parse id as an integer with the following error: '%s'.",
			err.Error(),
		)
	}
	out := &URLForward{
		ID:        &id,
		Subdomain: f.Subdomain,
		Location:  f.Location,
		Type:      f.Type,
	}
	switch f.IncludePath {
	case "yes":
		out.IncludePath = true
	case "no":
		out.IncludePath = false
	default:
		return nil, fmt.Errorf(
			"Expected include path of 'yes' or 'no', got: '%s'.",
			f.IncludePath,
		)
	}
	switch f.Wildcard {
	case "yes":
		out.Wildcard = true
	case "no":
		out.Wildcard = false
	default:
		return nil, fmt.Errorf(
			"Expected wildcard of 'yes' or 'no', got: '%s'.",
			f.Wildcard,
		)
	}
	return out, nil
}

func (c *Client) URLForwards(
	ctx context.Context,
	domain string,
) (
	[]*URLForward,
	[]error,
) {
	path := "domain/getUrlForwarding/" + domain
	req := &Credentials{}
	var res struct {
		Status
		Forwards []urlforward `json:"forwards"`
	}
	err := c.post(ctx, path, req, &res)
	if err != nil {
		return nil, []error{err}
	}
	forwards := []*URLForward{}
	errs := []error{}
	for _, forward := range res.Forwards {
		f, err := forward.convert()
		if err != nil {
			errs = append(errs, err)
		}
		forwards = append(forwards, f)
	}
	return forwards, errs
}

func (c *Client) AddURLForward(
	ctx context.Context,
	domain string,
	forward *URLForward,
) error {
	path := "domain/addUrlForward/" + domain
	req := &struct {
		Credentials
		urlforward
	}{urlforward: *forward.convert()}
	var res Status
	err := c.post(ctx, path, req, &res)
	return err
}

func (c *Client) DeleteURLForward(
	ctx context.Context,
	domain string,
	id int64,
) error {
	// ! Typo in the docs
	path := "domain/deleteUrlForward/" + domain + "/" + strconv.FormatInt(id, 10)
	req := &Credentials{}
	var res Status
	err := c.post(ctx, path, req, &res)
	return err
}
