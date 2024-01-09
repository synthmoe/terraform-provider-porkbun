// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package porkbun

type Status struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type status interface {
	getStatus() string
	getMessage() string
}

func (s *Status) getStatus() string {
	return s.Status
}

func (s *Status) getMessage() string {
	return s.Message
}
