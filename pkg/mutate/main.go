package mutate

import (
	corev1 "k8s.io/api/core/v1"
)

type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func AddVolume(pod *corev1.Pod) (patch []PatchOperation) {
	volume := corev1.Volume{
		Name: "elastic-apm-agent",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

	volumes := append(pod.Spec.Volumes, volume)
	patch = append(patch, PatchOperation{
		Op:    "replace",
		Path:  "/spec/volumes",
		Value: volumes,
	})

	return patch
}

func AddInitContainer(pod *corev1.Pod, initContainerImage string) (patch []PatchOperation) {
	initContainers := []corev1.Container{}

	initContainer := corev1.Container{
		Image: initContainerImage,
		Name:  "apm-agent-java",
		Command: []string{
			"cp",
			"-v",
			"/usr/agent/elastic-apm-agent.jar",
			"/elastic/apm/agent",
		},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      "elastic-apm-agent",
				MountPath: "/elastic/apm/agent",
			},
		},
	}

	initContainers = append(initContainers, initContainer)

	var op string
	if len(pod.Spec.InitContainers) != 0 {
		initContainers = append(initContainers, pod.Spec.InitContainers...)
		op = "replace"
	} else {
		op = "add"
	}

	patch = append(patch, PatchOperation{
		Op:    op,
		Path:  "/spec/initContainers",
		Value: initContainers,
	})

	return patch
}

func MutateContainers(pod *corev1.Pod, extraEnvVars []corev1.EnvVar) (patch []PatchOperation) {
	containers := []corev1.Container{}
	envVar := corev1.EnvVar{
		Name:  "JAVA_TOOL_OPTIONS",
		Value: "-javaagent:/elastic/apm/agent/elastic-apm-agent.jar",
	}
	volumeMount := corev1.VolumeMount{
		Name:      "elastic-apm-agent",
		MountPath: "/elastic/apm/agent",
	}

	for _, v := range pod.Spec.Containers {
		v.Env = append(v.Env, envVar)
		for _, envVar := range extraEnvVars {
			v.Env = append(v.Env, envVar)
		}
		v.VolumeMounts = append(v.VolumeMounts, volumeMount)
		containers = append(containers, v)
		patch = append(patch, PatchOperation{
			Op:    "replace",
			Path:  "/spec/containers",
			Value: containers,
		})
	}

	return patch
}
