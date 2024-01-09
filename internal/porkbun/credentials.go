// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package porkbun

type Credentials struct {
	APIKey       string `json:"apikey"`
	SecretAPIKey string `json:"secretapikey"`
}

type credentials interface {
	setAPIKey(string)
	setSecretAPIKey(string)
}

func (c *Credentials) setAPIKey(apiKey string) {
	c.APIKey = apiKey
}

func (c *Credentials) setSecretAPIKey(secretAPIKey string) {
	c.SecretAPIKey = secretAPIKey
}
