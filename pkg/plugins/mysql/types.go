// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mysql

type MySQLOptions struct {
	Host             string      `validate:"required"`
	Port             string      `validate:"required"`
	DatabaseName     string      `validate:"required"`
	DatabaseUsername string      `validate:"required"`
	DatabasePassword string      `validate:"required"`
	SSL              *SSLOptions `validate:"required"`
	SSH              *SSHOptions `validate:"required"`
}

type SSLOptions struct {
	SSL        bool
	ServerCert string `validate:"required_if=SSL true"`
	ClientKey  string `validate:"required_if=SSL true"`
	ClientCert string `validate:"required_if=SSL true"`
}

type SSHOptions struct {
	SSH           bool
	SSHHost       string `validate:"required_if=SSH true"`
	SSHPort       string `validate:"required_if=SSH true"`
	SSHUsername   string `validate:"required_if=SSH true"`
	SSHPassword   string `validate:"required_if=SSH true"`
	SSHPrivateKey string `validate:"required_if=SSH true"`
	SSHPassphrase string `validate:"required_if=SSH true"`
}

type MySQLQuery struct {
	Mode  string `validate:"required,oneof=gui sql"`
	Query string `validate:"required"`
}