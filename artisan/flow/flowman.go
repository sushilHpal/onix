/*
  Onix Config Manager - Artisan
  Copyright (c) 2018-2021 by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/
package flow

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/gatblau/onix/artisan/core"
	"github.com/gatblau/onix/artisan/crypto"
	"github.com/gatblau/onix/artisan/data"
	"github.com/gatblau/onix/artisan/registry"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// the pipeline generator requires at least the flow definition
// if a build file is passed then step variables can be inferred from it
type Manager struct {
	flow         *Flow
	buildFile    *data.BuildFile
	bareFlowPath string
}

func NewFromPath(bareFlowPath, buildPath string) (*Manager, error) {
	// check the flow path to see if bare flow is named correctly
	if !strings.HasSuffix(bareFlowPath, "_bare.yaml") {
		core.RaiseErr("a bare flow is required, the naming convention is [flow_name]_bare.yaml")
	}
	m := &Manager{
		bareFlowPath: bareFlowPath,
	}
	flow, err := LoadFlow(bareFlowPath, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot load flow definition from %s: %s", bareFlowPath, err)
	}
	m.flow = flow
	// if a build file is defined, then load it
	if len(buildPath) > 0 {
		buildPath = core.ToAbs(buildPath)
		buildFile, err := data.LoadBuildFile(path.Join(buildPath, "build.yaml"))
		if err != nil {
			return nil, fmt.Errorf("cannot load build file from %s: %s", buildPath, err)
		}
		m.buildFile = buildFile
	}
	err = m.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid generator: %s", err)
	}
	return m, nil
}

func (m *Manager) Merge() error {
	local := registry.NewLocalRegistry()
	if m.flow.RequiresSource() {
		if m.buildFile == nil {
			return fmt.Errorf("a build.yaml file is required to fill the flow")
		}
		if len(m.buildFile.GitURI) == 0 {
			return fmt.Errorf("a 'git_uri' is required in the build.yaml")
		}
		m.flow.GitURI = m.buildFile.GitURI
		m.flow.AppIcon = m.buildFile.AppIcon
	}
	for _, step := range m.flow.Steps {
		if len(step.Package) > 0 {
			name, err := core.ParseName(step.Package)
			core.CheckErr(err, "invalid step %s package name %s", step.Name, step.Package)
			// get the package manifest
			manifest := local.GetManifest(name)
			step.Input = data.SurveyInputFromManifest(name, step.Function, manifest, true)
			// collects credentials to retrieve package from registry
			m.surveyRegistryCreds(step.Package)
		} else {
			// if the step has a function
			if len(step.Function) > 0 {
				// add exported inputs to the step
				step.Input = data.SurveyInputFromBuildFile(step.Function, m.buildFile, true)
			} else {
				// read input from from runtime_uri
				step.Input = data.SurveyInputFromURI(step.RuntimeManifest, true)
			}
		}
	}
	return nil
}

func (m *Manager) surveyRegistryCreds(packageName string) {
	name, _ := core.ParseName(packageName)
	// if the credentials for the package domain have not been added
	if !m.flow.HasDomain(name.Domain) {
		var user, pwd string
		// prompt for the registry username
		userPrompt := &survey.Password{
			Message: fmt.Sprintf("secret => REGISTRY USER (for %s):", packageName),
		}
		core.HandleCtrlC(survey.AskOne(userPrompt, &user, survey.WithValidator(survey.Required)))

		// prompt for the registry password
		pwdPrompt := &survey.Password{
			Message: fmt.Sprintf("secret => REGISTRY PASSWORD (for %s):", packageName),
		}
		core.HandleCtrlC(survey.AskOne(pwdPrompt, &pwd, survey.WithValidator(survey.Required)))

		// add the credentials to the flow list
		m.flow.Credential = append(m.flow.Credential, &Credential{
			User:     user,
			Password: pwd,
			Domain:   name.Domain,
		})
	}
}

func (m *Manager) YamlString() (string, error) {
	b, err := yaml.Marshal(m.flow)
	if err != nil {
		return "", fmt.Errorf("cannot marshal execution flow: %s", err)
	}
	return string(b), nil
}

func LoadFlow(path string, key *crypto.PGP) (*Flow, error) {
	var err error
	if len(path) == 0 {
		return nil, fmt.Errorf("flow definition is required")
	}
	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("cannot get absolute path for %s: %s", path, err)
		}
	}
	flowBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read flow definition %s: %s", path, err)
	}
	if key != nil {
		flowBytes, err = key.Decrypt(flowBytes)
		if err != nil {
			return nil, fmt.Errorf("cannot decrypt flow %s: %s", path, err)
		}
	}
	flow := new(Flow)
	err = yaml.Unmarshal(flowBytes, flow)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal flow definition %s: %s", path, err)
	}
	return flow, nil
}

func (m *Manager) validate() error {
	// check that the steps have the required attributes set
	for _, step := range m.flow.Steps {
		if len(step.Runtime) == 0 {
			return fmt.Errorf("invalid step %s, runtime is missing", step.Name)
		}
	}
	return nil
}

func (m *Manager) Save() error {
	y, err := yaml.Marshal(m.flow)
	if err != nil {
		return fmt.Errorf("cannot marshal bare flow: %s", err)
	}
	err = ioutil.WriteFile(m.path(), y, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot save merged flow: %s", err)
	}
	return nil
}

// get the merged flow path
func (m *Manager) path() string {
	dir, file := filepath.Split(m.bareFlowPath)
	filename := core.FilenameWithoutExtension(file)
	return filepath.Join(dir, fmt.Sprintf("%s.yaml", filename[0:len(filename)-len("_bare")]))
}
