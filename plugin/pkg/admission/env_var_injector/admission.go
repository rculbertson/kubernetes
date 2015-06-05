package set_env_var

import (
	"io"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/admission"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"

)

func init() {
	admission.RegisterPlugin("EnvVarInjector", func(client client.Interface, config io.Reader) (admission.Interface, error) {
		return NewEnvVarInjector(), nil
	})
}

type envVarInjector struct{}

func (envVarInjector) Admit(attr admission.Attributes) (err error) {
	switch attr.GetKind() {
	case "Pod":
		pod := attr.GetObject().(*api.Pod)
		return handlePod(pod)
	case "ReplicationController":
		pod := attr.GetObject().(*api.ReplicationController)
		return handleReplicationController(pod)
	}
	return
}

func handleReplicationController(controller *api.ReplicationController) (err error) {
	return handlePodSpec(&controller.Spec.Template.Spec)
}

func handlePod(pod *api.Pod) (err error) {
	return handlePodSpec(&pod.Spec)
}

func handlePodSpec(podSpec *api.PodSpec) (err error) {
	// TODO: List of environment variables should be read from a config file.
	// For now we'll inject SPOTIFY_DOMAIN only.
	defaultEnvVar := api.EnvVar{Name:"SPOTIFY_DOMAIN", Value:"gcb2"}

	containers := podSpec.Containers

	// iterate over all containers in the pod
	for i := 0; i < len(containers); i++ {
		overridePresent := false

		// check if this config already contains the environment variable we're about to set
		for _, envVar := range containers[i].Env {
			if envVar.Name == defaultEnvVar.Name {
				overridePresent = true
				break
			}
		}

		// set the default value if an override value is not set
		if !overridePresent {
			containers[i].Env = append(containers[i].Env, defaultEnvVar)
		}
	}

	return
}

func (envVarInjector) Handles(operation admission.Operation) bool {
	return true
}

func NewEnvVarInjector() admission.Interface {
	return new(envVarInjector)
}