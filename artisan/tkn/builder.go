package tkn

/*
  Onix Config Manager - Artisan
  Copyright (c) 2018-Present by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/
import (
	"bytes"
	"fmt"
	"github.com/gatblau/onix/artisan/crypto"
	"github.com/gatblau/onix/artisan/flow"
	"strings"
	"time"
)

const (
	ApiVersionSecret        = "v1"
	ApiVersionRoute         = "route.openshift.io/v1"
	ApiVersionTekton        = "tekton.dev/v1alpha1"
	ApiVersionTektonTrigger = "triggers.tekton.dev/v1alpha1"

	// ServiceAccountName account below is created by the tekton operator in OpenShift but has to be created manually
	// if a plain Kubernetes version is used
	ServiceAccountName = "pipeline"
)

// Builder tekton builder
type Builder struct {
	flow *flow.Flow
}

// NewBuilder creates a new tekton builder
func NewBuilder(flow *flow.Flow) *Builder {
	return &Builder{
		flow: flow,
	}
}

// BuildBuffer creates a buffer with all K8S resources required to create a tekton pipeline out of an Artisan flow
func (b *Builder) BuildBuffer() bytes.Buffer {
	buffer := bytes.Buffer{}
	resx, _, _ := b.Build()
	for _, resource := range resx {
		buffer.Write(resource)
		buffer.WriteString("\n---\n")
	}
	return buffer
}

// Build creates a slice with all K8S resources required to create a tekton pipleine out of an Artisan flow
func (b *Builder) Build() ([][]byte, string, bool) {
	result := make([][]byte, 0)
	// writes a task
	task := b.newTask()
	result = append(result, ToYaml(task, "Task"))
	// write secrets with credentials
	secrets := b.newCredentialsSecret()
	if secrets != nil {
		result = append(result, ToYaml(secrets, "Secret"))
	}
	// write secrets with keys
	keysSecret := b.newKeySecrets()
	if keysSecret != nil {
		result = append(result, ToYaml(keysSecret, "Keys Secret"))
	}
	// write secrets with files
	filesSecret := b.newFileSecrets()
	if filesSecret != nil {
		result = append(result, ToYaml(filesSecret, "Files Secret"))
	}
	// write pipeline
	pipeline := b.newPipeline()
	result = append(result, ToYaml(pipeline, "Pipeline"))

	pipelineRun := b.newPipelineRun()

	// if source code repository is required by the pipeline
	if b.flow.RequiresGitSource() {
		// need to add git repo in resources of pipeline run
		pipelineRun.Spec.Resources = []*Resources{
			{
				Name: b.codeRepoResourceName(b.flow.Name),
				ResourceRef: &ResourceRef{
					Name: b.codeRepoResourceName(b.flow.Name),
				},
			},
		}
		// add the following resources:
		// tekton pipeline resource
		pipelineResource := b.newPipelineResource()
		result = append(result, ToYaml(pipelineResource, "PipelineResource"))

		// NOTE: resources below are OpenShift specific so removing so it is compatible across all K8S implementations
		//
		// // tekton event listener
		// eventListener := b.newEventListener()
		// result = append(result, ToYaml(eventListener, "EventListener"))
		//
		// // k8s route
		// route := b.newRoute()
		// result = append(result, ToYaml(route, "Route"))
		//
		// // tekton trigger binding
		// triggerBinding := b.newTriggerBinding()
		// result = append(result, ToYaml(triggerBinding, "TriggerBinding"))
		//
		// // tekton trigger template
		// triggerTemplate := b.newTriggerTemplate()
		// result = append(result, ToYaml(triggerTemplate, "TriggerTemplate"))
	}
	result = append(result, ToYaml(pipelineRun, "Pipeline Run"))
	return result, pipelineRun.Metadata.Name, b.flow.RequiresGitSource()
}

// task
func (b *Builder) newTask() *Task {
	t := new(Task)
	t.APIVersion = ApiVersionTekton
	t.Kind = "Task"
	t.Metadata = &Metadata{
		Name:      b.buildTaskName(b.flow.Name),
		Namespace: b.namespace(),
	}
	t.Spec = &TaskSpec{
		Steps:   b.newSteps(),
		Volumes: b.newVolumes(),
	}
	if b.flow.RequiresGitSource() {
		t.Spec.Resources = &TaskResources{
			Inputs: []*Inputs{
				{
					Name: "source",
					Type: "git",
				},
			},
		}
	}
	return t
}

func (b *Builder) newSteps() []*Steps {
	var steps = make([]*Steps, 0)
	for _, step := range b.flow.Steps {
		s := &Steps{
			Name:       step.Name,
			Image:      step.Runtime,
			WorkingDir: "/workspace/source",
		}
		// if the step is marked a privileged in the Artisan flow
		// adds a security context to override the non-privileged setting of the pipeline run
		if step.Privileged {
			s.SecurityContext = &StepSecurityContext{
				Privileged:   true,
				RunAsNonRoot: false,
				RunAsUser:    0,
			}
		}
		// if the step requires keys
		if step.Input != nil {
			if len(step.Input.Key) > 0 {
				// add a volume mount for the keys
				s.VolumeMounts = []*VolumeMounts{
					{
						Name:      "keys-volume",
						MountPath: "/keys",
					},
				}
			}
			// if the step has vars or secrets or keys
			if len(step.Input.Var)+len(step.Input.Secret)+len(step.Input.Key) > 0 {
				// add to env
				s.Env = b.getEnv(step)
			}
			if len(step.Input.File) > 0 {
				// add a volume mount for the files
				if s.VolumeMounts != nil {
					s.VolumeMounts = append(s.VolumeMounts, &VolumeMounts{
						Name:      "files-volume",
						MountPath: "/files",
					})
				} else {
					s.VolumeMounts = []*VolumeMounts{
						{
							Name:      "files-volume",
							MountPath: "/files",
						},
					}
				}
			}
		}
		// add the environment information required by the Artisan runtime to work
		// see here: https://github.com/gatblau/artisan/tree/master/runtime
		s.Env = b.addRuntimeInterfaceVars(b.flow.Name, step, s.Env)
		steps = append(steps, s)
	}
	return steps
}

func (b *Builder) getEnv(step *flow.Step) []*Env {
	env := make([]*Env, 0)
	// if there is an input defined
	if step.Input != nil {
		// add variables
		for _, variable := range step.Input.Var {
			env = append(env, &Env{
				Name:  variable.Name,
				Value: variable.Value,
			})
		}
		// add secrets
		for _, secret := range step.Input.Secret {
			if strings.HasSuffix(secret.Name, "OXART_REG_USER") {
				// if the secret is a art reg username, convert it to the format expected by the runtime
				env = append(env, &Env{
					Name: "OXART_REG_USER",
					ValueFrom: &ValueFrom{
						SecretKeyRef: &SecretKeyRef{
							Name: b.secretName(),
							Key:  secret.Name,
						}},
				})
			} else if strings.HasSuffix(secret.Name, "OXART_REG_PWD") {
				// if the secret is a art reg username, convert it to the format expected by the runtime
				env = append(env, &Env{
					Name: "OXART_REG_PWD",
					ValueFrom: &ValueFrom{
						SecretKeyRef: &SecretKeyRef{
							Name: b.secretName(),
							Key:  secret.Name,
						}},
				})
			} else {
				// an ordinary secret
				env = append(env, &Env{
					Name: secret.Name,
					ValueFrom: &ValueFrom{
						SecretKeyRef: &SecretKeyRef{
							Name: b.secretName(),
							Key:  secret.Name,
						}},
				})
			}
		}
		// add keys
		for _, key := range step.Input.Key {
			env = append(env, &Env{
				Name:  key.Name,
				Value: key.Path,
			})
		}
	}
	return env
}

func (b *Builder) addRuntimeInterfaceVars(flowName string, step *flow.Step, env []*Env) []*Env {
	if len(step.Function) > 0 {
		env = append(env, &Env{
			Name:  "OXART_FX_NAME",
			Value: step.Function,
		})
	}
	if len(step.Package) > 0 {
		env = append(env, &Env{
			Name:  "OXART_PACKAGE_NAME",
			Value: step.Package,
		})
		if len(step.PackageSource) > 0 {
			env = append(env, &Env{
				Name:  "OXART_PACKAGE_SOURCE",
				Value: step.PackageSource,
			})
		}
	}
	return env
}

func (b *Builder) newVolumes() []*Volumes {
	var vols []*Volumes = nil
	if b.flow.RequiresKey() {
		if vols == nil {
			vols = make([]*Volumes, 0)
		}
		vols = append(vols, &Volumes{
			Name: "keys-volume",
			Secret: &Secret{
				SecretName: b.keysSecretName(),
			},
		})
	}
	if b.flow.RequiresFile() {
		if vols == nil {
			vols = make([]*Volumes, 0)
		}
		vols = append(vols, &Volumes{
			Name: "files-volume",
			Secret: &Secret{
				SecretName: b.filesSecretName(),
			},
		})
	}
	return vols
}

func (b *Builder) newCredentialsSecret() *Secret {
	if b.flow.RequiresSecrets() {
		s := new(Secret)
		s.APIVersion = ApiVersionSecret
		s.Kind = "Secret"
		s.Type = "Opaque"
		s.Metadata = &Metadata{
			Name:      b.secretName(),
			Namespace: b.namespace(),
		}
		credentials := make(map[string]string)
		for _, step := range b.flow.Steps {
			if step.Input != nil && step.Input.Secret != nil {
				for _, secret := range step.Input.Secret {
					name := secret.Name
					credentials[name] = secret.Value
				}
			}
		}
		// add flow level secrets
		if b.flow.Input != nil && b.flow.Input.Secret != nil {
			for _, cred := range b.flow.Input.Secret {
				credentials[cred.Name] = cred.Value
				credentials[cred.Name] = cred.Value
			}
		}
		s.StringData = &credentials
		return s
	}
	return nil
}

func (b *Builder) newKeySecrets() *Secret {
	if b.flow.RequiresKey() {
		s := new(Secret)
		s.APIVersion = ApiVersionSecret
		s.Kind = "Secret"
		s.Type = "Opaque"
		s.Metadata = &Metadata{
			Name:      b.keysSecretName(),
			Namespace: b.namespace(),
		}
		keysDict := make(map[string]string)
		var name string
		for _, step := range b.flow.Steps {
			if step.Input != nil {
				keys := step.Input.Key
				for _, key := range keys {
					prefix := crypto.KeyNamePrefix(key.PackageGroup, key.PackageName)
					if key.Private {
						name = crypto.PrivateKeyName(prefix, "pgp")
					} else {
						name = crypto.PublicKeyName(prefix, "pgp")
					}
					keysDict[name] = key.Value
				}
			}
		}
		s.StringData = &keysDict
		return s
	}
	return nil
}

func (b *Builder) newFileSecrets() *Secret {
	if b.flow.RequiresFile() {
		s := new(Secret)
		s.APIVersion = ApiVersionSecret
		s.Kind = "Secret"
		s.Type = "Opaque"
		s.Metadata = &Metadata{
			Name:      b.filesSecretName(),
			Namespace: b.namespace(),
		}
		filesDict := make(map[string]string)
		for _, step := range b.flow.Steps {
			if step.Input != nil {
				files := step.Input.File
				for _, file := range files {
					filesDict[file.Path] = file.Content
				}
			}
		}
		s.StringData = &filesDict
		return s
	}
	return nil
}

// pipeline
func (b *Builder) newPipeline() *Pipeline {
	p := new(Pipeline)
	p.Kind = "Pipeline"
	p.APIVersion = ApiVersionTekton
	p.Metadata = &Metadata{
		Name:      b.pipelineName(b.flow.Name),
		Namespace: b.namespace(),
	}
	var (
		inputs    []*Inputs
		resources []*Resources
	)
	if b.flow.RequiresGitSource() {
		inputs = []*Inputs{
			{
				Name:     "source",
				Resource: b.codeRepoResourceName(b.flow.Name),
			},
		}
		resources = []*Resources{
			{
				Name: b.codeRepoResourceName(b.flow.Name),
				Type: "git",
			},
		}
	}
	p.Spec = &Spec{
		Resources: resources,
		Params: []*Params{
			{
				Name:        "deployment-name",
				Type:        "string",
				Description: "the unique name for this deployment",
			},
		},
		Tasks: []*Tasks{
			{
				Name: b.buildTaskName(b.flow.Name),
				TaskRef: &TaskRef{
					Name: b.buildTaskName(b.flow.Name),
				},
				Resources: &Resources{
					Inputs: inputs,
				},
			},
		},
	}
	return p
}

// pipeline resource
func (b *Builder) newPipelineResource() *PipelineResource {
	r := new(PipelineResource)
	r.APIVersion = ApiVersionTekton
	r.Kind = "PipelineResource"
	r.Metadata = &Metadata{
		Name:      b.codeRepoResourceName(b.flow.Name),
		Namespace: b.namespace(),
	}
	r.Spec = &Spec{
		Type: "git",
		Params: []*Params{
			{
				Name:  "url",
				Value: b.flow.GitURI,
			},
		},
	}
	return r
}

// pipeline run
func (b *Builder) newPipelineRun() *PipelineRun {
	return b.NewNamedPipelineRun(b.flow.Name, b.namespace())
}

// NewNamedPipelineRun create a pipeline run for the passed in name
func (b *Builder) NewNamedPipelineRun(flowName, namespace string) *PipelineRun {
	r := new(PipelineRun)
	r.Kind = "PipelineRun"
	r.APIVersion = ApiVersionTekton
	r.Spec = &Spec{
		// this is the default service account created by the Tekton operator
		ServiceAccountName: "pipeline",
		PipelineRef: &PipelineRef{
			Name: b.pipelineName(flowName),
		},
		Params: []*Params{
			{
				Name:  "deployment-name",
				Value: flowName,
			},
		},
		// always run the pipeline as non root user using the Artisan user Id for the runtimes (i.e. 100000000)
		// this prevents the pipeline to run as root user in plain Kubernetes
		// Artisan runtimes cannot run as root
		PodTemplate: &PodTemplate{
			SecurityContext: &PipelineSecurityContext{
				RunAsNonRoot: true,
				FsGroup:      100000000,
				RunAsUser:    100000000,
			},
		},
	}
	r.Metadata = &Metadata{
		Name:      b.pipelineRunName(flowName),
		Namespace: namespace,
	}
	return r
}

// event listener
func (b *Builder) newEventListener() *EventListener {
	e := new(EventListener)
	e.APIVersion = ApiVersionTektonTrigger
	e.Kind = "EventListener"
	e.Metadata = &Metadata{
		Name:      encode(b.flow.Name),
		Namespace: b.namespace(),
		Labels: &Labels{
			AppOpenshiftIoRuntime: b.flow.AppIcon,
		},
	}
	e.Spec = &Spec{
		ServiceAccountName: ServiceAccountName,
		Triggers: []*Triggers{
			{
				Bindings: []*Bindings{
					{
						Name: encode(b.flow.Name),
					},
				},
				Template: &Template{
					Name: encode(b.flow.Name),
				},
			},
		},
	}
	return e
}

// route
func (b *Builder) newRoute() *Route {
	r := new(Route)
	r.APIVersion = ApiVersionRoute
	r.Kind = "Route"
	r.Metadata = &Metadata{
		Name:      fmt.Sprintf("el-%s", encode(b.flow.Name)),
		Namespace: b.namespace(),
		Labels: &Labels{
			Application: fmt.Sprintf("%s-https", encode(b.flow.Name)),
		},
		Annotations: &Annotations{
			Description: "Route for the Pipeline Event Listener.",
		},
	}
	r.Spec = &Spec{
		Port: &Port{
			TargetPort: "8080",
		},
		TLS: &TLS{
			InsecureEdgeTerminationPolicy: "Redirect",
			Termination:                   "edge",
		},
		To: &To{
			Kind: "Service",
			Name: fmt.Sprintf("el-%s", encode(b.flow.Name)),
		},
	}
	return r
}

// trigger binding
func (b *Builder) newTriggerBinding() *TriggerBinding {
	t := new(TriggerBinding)
	t.APIVersion = ApiVersionTektonTrigger
	t.Kind = "TriggerBinding"
	t.Metadata = &Metadata{
		Name:      encode(b.flow.Name),
		Namespace: b.namespace(),
	}
	t.Spec = &Spec{
		Params: []*Params{
			{
				Name:  "git-repo-url",
				Value: "$(body.project.web_url)",
			},
			{
				Name:  "git-repo-name",
				Value: "$(body.repository.name)",
			},
			{
				Name:  "git-revision",
				Value: "$(body.commits[0].id)",
			},
		},
	}
	return t
}

// trigger template
func (b *Builder) newTriggerTemplate() *PipelineRun {
	pipeResx := b.newPipelineResourceTriggerTemplate()
	pipeRun := b.newPipelineRunTriggerTemplate()

	t := new(PipelineRun)
	t.APIVersion = ApiVersionTektonTrigger
	t.Kind = "TriggerTemplate"
	t.Metadata = &Metadata{
		Name:      encode(b.flow.Name),
		Namespace: b.namespace(),
	}
	t.Spec = &Spec{
		Params: []*Params{
			{
				Name:        "git-repo-url",
				Description: "The git repository url",
			},
			{
				Name:        "git-repo-name",
				Description: "The git repository name",
			},
			{
				Name:        "git-revision",
				Description: "The git revision",
				Default:     "master",
			},
		},
		ResourceTemplates: []interface{}{pipeResx, pipeRun},
	}
	return t
}

func (b *Builder) newPipelineResourceTriggerTemplate() *PipelineResource {
	r := new(PipelineResource)
	r.APIVersion = ApiVersionTekton
	r.Kind = "PipelineResource"
	r.Metadata = &Metadata{
		Name:      "$(params.git-repo-name)-git-repo-$(uid)",
		Namespace: b.namespace(),
	}
	r.Spec = &Spec{
		Type: "git",
		Params: []*Params{
			{
				Name:  "revision",
				Value: "$(params.git-revision)",
			},
			{
				Name:  "url",
				Value: b.flow.GitURI,
			},
		},
	}
	return r
}

func (b *Builder) newPipelineRunTriggerTemplate() *PipelineRun {
	r := new(PipelineRun)
	r.Kind = "PipelineRun"
	r.APIVersion = ApiVersionTekton
	r.Metadata = &Metadata{
		Name:      "$(params.git-repo-name)-app-pr-$(uid)",
		Namespace: b.namespace(),
	}
	r.Spec = &Spec{
		ServiceAccountName: ServiceAccountName,
		PipelineRef: &PipelineRef{
			Name: b.pipelineName(b.flow.Name),
		},
		Resources: []*Resources{
			{
				Name: b.codeRepoResourceName(b.flow.Name),
				ResourceRef: &ResourceRef{
					Name: "$(params.git-repo-name)-git-repo-$(uid)",
				},
			},
		},
		Params: []*Params{
			{
				Name:  "deployment-name",
				Value: "$(params.git-repo-name)",
			},
		},
	}
	return r
}

// return the name of the application build task
func (b *Builder) buildTaskName(flowName string) string {
	return fmt.Sprintf("%s-build-task", encode(flowName))
}

// return the name of the code repository resource
func (b *Builder) codeRepoResourceName(flowName string) string {
	return fmt.Sprintf("%s-code-repo", encode(flowName))
}

// return the name of the code repository resource
func (b *Builder) pipelineName(flowName string) string {
	return fmt.Sprintf("%s-pl", encode(flowName))
}

// return the name of the code repository resource
func (b *Builder) pipelineRunName(flowName string) string {
	return fmt.Sprintf("%s-pr-%d", encode(flowName), time.Now().Nanosecond())
}

// return the name of the code repository resource
func (b *Builder) secretName() string {
	return fmt.Sprintf("%s-creds-secret", encode(b.flow.Name))
}

func (b *Builder) keysSecretName() string {
	return fmt.Sprintf("%s-keys-secret", encode(b.flow.Name))
}

func (b *Builder) filesSecretName() string {
	return fmt.Sprintf("%s-files-secret", encode(b.flow.Name))
}

// retrieves the namespace label in the flow
func (b *Builder) namespace() string {
	return strings.Trim(b.flow.Labels["namespace"], " ")
}
