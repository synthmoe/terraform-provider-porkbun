// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package porkbun

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type DNSRecord struct {
	ID        *int64
	Subdomain string
	Type      string
	Content   string
	TTL       int64
	Priority  *int64
	Notes     string
}

type dnsrecord struct {
	ID        json.Number `json:"id,omitempty"`
	Subdomain string      `json:"name"`
	Type      string      `json:"type"`
	Content   string      `json:"content"`
	TTL       json.Number `json:"ttl"`
	Priority  json.Number `json:"prio"`
	Notes     string      `json:"notes"`
}

func (r *DNSRecord) convert() *dnsrecord {
	id := ""
	if r.ID != nil {
		id = strconv.FormatInt(*r.ID, 10)
	}

	priority := ""
	if r.Priority != nil {
		priority = strconv.FormatInt(*r.Priority, 10)
	}

	return &dnsrecord{
		ID:        json.Number(id),
		Subdomain: r.Subdomain,
		Type:      r.Type,
		Content:   r.Content,
		TTL:       json.Number(strconv.FormatInt(r.TTL, 10)),
		Priority:  json.Number(priority),
		Notes:     r.Notes,
	}
}

func (r *dnsrecord) convert() (*DNSRecord, error) {
	record := &DNSRecord{
		Subdomain: r.Subdomain,
		Type:      r.Type,
		Content:   r.Content,
		Notes:     r.Notes,
	}

	if r.ID.String() == "" {
		id, err := r.ID.Int64()
		if err != nil {
			return nil, fmt.Errorf(
				"Failed to parse id as an integer with the following error: '%s'.",
				err.Error(),
			)
		}
		*record.ID = id
	}

	ttl, err := r.TTL.Int64()
	if err != nil {
		return nil, fmt.Errorf(
			"Failed to parse ttl as an integer with the following error: '%s'.",
			err.Error(),
		)
	}
	record.TTL = ttl

	if r.Priority.String() != "" {
		priority, err := r.Priority.Int64()
		if err != nil {
			return nil, fmt.Errorf(
				"Failed to parse priority as an integer with the following error: '%s'.",
				err.Error(),
			)
		}
		record.Priority = &priority
	}

	return record, nil
}

func (c *Client) DNSRecords(
	ctx context.Context,
	domain string,
	id *int64,
) (
	[]*DNSRecord,
	[]error,
) {
	path := "dns/retrieve/" + domain
	if id != nil {
		path += "/" + strconv.FormatInt(*id, 10)
	}
	req := &Credentials{}
	var res struct {
		Status
		Records []dnsrecord `json:"records"`
	}
	err := c.post(ctx, path, req, &res)
	if err != nil {
		return nil, []error{err}
	}
	records := []*DNSRecord{}
	errs := []error{}
	for _, record := range res.Records {
		r, err := record.convert()
		if err != nil {
			errs = append(errs, err)
		}
		records = append(records, r)
	}
	return records, errs
}

func (c *Client) DNSRecordsByTypeName(
	ctx context.Context,
	domain string,
	type_ string,
	subdomain string,
) (
	[]*DNSRecord,
	[]error,
) {
	path := "dns/retrieveByNameType/" + domain + "/" + type_
	if subdomain != "" {
		path += "/" + subdomain
	}
	req := &Credentials{}
	var res struct {
		Status
		Records []dnsrecord `json:"records"`
	}
	err := c.post(ctx, path, req, &res)
	if err != nil {
		return nil, []error{err}
	}
	records := []*DNSRecord{}
	errs := []error{}
	for _, record := range res.Records {
		r, err := record.convert()
		if err != nil {
			errs = append(errs, err)
		}
		records = append(records, r)
	}
	return records, errs
}

func (c *Client) CreateDNSRecord(
	ctx context.Context,
	domain string,
	record *DNSRecord,
) (
	int64,
	error,
) {
	path := "dns/create/" + domain
	req := &struct {
		Credentials
		dnsrecord
	}{dnsrecord: *record.convert()}
	var res struct {
		Status
		ID int64 `json:"id"`
	}
	err := c.post(ctx, path, req, &res)
	return res.ID, err
}

func (c *Client) EditDNSRecord(
	ctx context.Context,
	domain string,
	record *DNSRecord,
) error {
	path := "dns/edit/" + domain + "/" + strconv.FormatInt(*record.ID, 10)
	req := &struct {
		Credentials
		dnsrecord
	}{dnsrecord: *record.convert()}
	var res Status
	err := c.post(ctx, path, req, &res)
	return err
}

func (c *Client) DeleteDNSRecord(
	ctx context.Context,
	domain string,
	id int64,
) error {
	path := "dns/delete/" + domain + "/" + strconv.FormatInt(id, 10)
	req := &Credentials{}
	var res Status
	err := c.post(ctx, path, req, &res)
	return err
}
