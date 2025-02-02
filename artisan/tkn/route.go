/*
  Onix Config Manager - Artisan
  Copyright (c) 2018-Present by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/
package tkn

type Route struct {
	APIVersion string    `yaml:"apiVersion,omitempty"`
	Kind       string    `yaml:"kind,omitempty"`
	Metadata   *Metadata `yaml:"metadata,omitempty"`
	Spec       *Spec     `yaml:"spec,omitempty"`
}

type Annotations struct {
	Description string `yaml:"description,omitempty"`
}

type Port struct {
	TargetPort string `yaml:"targetPort,omitempty"`
}

type TLS struct {
	InsecureEdgeTerminationPolicy string `yaml:"insecureEdgeTerminationPolicy,omitempty"`
	Termination                   string `yaml:"termination,omitempty"`
}

type To struct {
	Kind string `yaml:"kind,omitempty"`
	Name string `yaml:"name,omitempty"`
}
