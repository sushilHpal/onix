/*
  Onix Config Manager - Artisan
  Copyright (c) 2018-Present by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package app

import (
	"fmt"
	"strconv"
)

// SvcManifest the manifest that describes how a software service should be configured
type SvcManifest struct {
	// Name the name of the service
	Name string `yaml:"name"`
	// Description describes what the service is all about
	Description string `yaml:"description"`
	// the port used by the http service
	Port interface{} `yaml:"port"`
	// the URI to determine if the service is ready to use
	ReadyURI string `yaml:"ready_uri,omitempty"`
	// the variables passed to the service (either ordinary or secret)
	Var []Var `yaml:"var,omitempty"`
	// the files used by the service (either ordinary or secret)
	File []File `yaml:"file,omitempty"`
	// one or more persistent volumes
	Volume []Volume `yaml:"volume,omitempty"`
	// a list of initialisation blocks for different builders
	Init []Init `yaml:"init"`
	// a list of command definitions that can be used in any of the initialisation blocks
	Script []Script `yaml:"scripts"`
	// the database configuration
	Db *Db `yaml:"db,omitempty"`
}

// PortMap return a parsed map of ports for the port attribute
func (m *SvcManifest) PortMap() (map[string]int, error) {
	ports := map[string]int{}
	if p, isString := m.Port.(string); isString {
		value, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		ports = map[string]int{
			"default": value,
		}
	} else if p, isMap := m.Port.(map[interface{}]interface{}); isMap {
		for key, value := range p {
			iv, err := strconv.Atoi(value.(string))
			if err != nil {
				return nil, err
			}
			ports[key.(string)] = iv
		}
	} else {
		return nil, fmt.Errorf("invalid port value: %s", m.Port)
	}
	return ports, nil
}

func (m SvcManifest) ScriptIx(scriptName string) int {
	for i, s := range m.Script {
		if s.Name == scriptName {
			return i
		}
	}
	return -1
}

// Init initialisation block containing commands to be run for a specific builder
type Init struct {
	Builder string   `yaml:"builder"`
	Scripts []string `yaml:"scripts"`
}

// Script defines a script that can be run in a host or a runtime
type Script struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Content     string `yaml:"content"`
	Runtime     string `yaml:"runtime,omitempty"`
}

// Var describes a variable used by a service
type Var struct {
	// the variable name
	Name string `yaml:"name"`
	// a human-readable description for the variable
	Description string `yaml:"description,omitempty"`
	// if defined, the fix value for the variable
	Value string `yaml:"value,omitempty"`
	// whether the variable should be treated as a secret
	Secret bool `yaml:"secret,omitempty"`
	// a default value for the variable
	Default string `yaml:"default,omitempty"`
}

// File describes a file used by a service
type File struct {
	// the file path
	Path string `yaml:"path"`
	// a human-readable description for the file
	Description string `yaml:"description,omitempty"`
	// whether the file should be treated as a secret
	Secret bool `yaml:"secret,omitempty"`
	// the template to use to create the file
	Template string `yaml:"template,omitempty"`
	// the content of the file (can be the result of evaluating a template)
	Content string `yaml:"content,omitempty"`
}

type Volume struct {
	// the name of the volume
	Name string `yaml:"name"`
	// the volume use description
	Description string `yaml:"description,omitempty"`
	// the volume source path
	Path string `yaml:"path"`
}

type Db struct {
	Name       string `yaml:"name"`
	AppVersion string `yaml:"app_version"`
	Host       string `yaml:"host"`
	Provider   string `yaml:"provider"`
	Port       int    `yaml:"port"`
	User       string `yaml:"user"`
	Pwd        string `yaml:"pwd"`
	AdminUser  string `yaml:"admin_user"`
	AdminPwd   string `yaml:"admin_pwd"`
	SchemaURI  string `yaml:"schema_uri"`
}
