package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	v1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func mutate(body []byte, verbose bool) ([]byte, error) {
	if verbose {
		log.Printf("recv: %s\n", string(body))
	}

	admReview := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	var err error
	var pod *corev1.Pod

	responseBody := []byte{}
	ar := admReview.Request
	resp := v1beta1.AdmissionResponse{}

	if ar != nil {
		if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
			return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
		}

		resp.Allowed = true
		resp.UID = ar.UID
		pT := v1beta1.PatchTypeJSONPatch
		resp.PatchType = &pT
		p := []patchOperation{}

		p = append(p, addVolume(pod)...)
		p = append(p, addInitContainer(pod)...)
		p = append(p, mutateContainers(pod)...)

		resp.Patch, err = json.Marshal(p)

		resp.Patch, err = json.Marshal(p)

		resp.Result = &metav1.Status{
			Status: "Success",
		}

		admReview.Response = &resp

		responseBody, err = json.Marshal(admReview)
		if err != nil {
			return nil, err
		}
	}

	if verbose {
		log.Printf("resp: %s\n", string(responseBody))
	}

	return responseBody, nil
}

func addVolume(pod *corev1.Pod) (patch []patchOperation) {
	volume := corev1.Volume{
		Name: "elastic-apm-agent",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

	volumes := append(pod.Spec.Volumes, volume)
	patch = append(patch, patchOperation{
		Op:    "replace",
		Path:  "/spec/volumes",
		Value: volumes,
	})

	return patch
}

func addInitContainer(pod *corev1.Pod) (patch []patchOperation) {
	initContainers := []corev1.Container{}

	initContainer := corev1.Container{
		Image: "docker.elastic.co/observability/apm-agent-java:1.12.0",
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

	patch = append(patch, patchOperation{
		Op:    op,
		Path:  "/spec/initContainers",
		Value: initContainers,
	})

	return patch
}

func mutateContainers(pod *corev1.Pod) (patch []patchOperation) {
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
		v.VolumeMounts = append(v.VolumeMounts, volumeMount)
		containers = append(containers, v)
		patch = append(patch, patchOperation{
			Op:    "replace",
			Path:  "/spec/containers",
			Value: containers,
		})
	}

	return patch
}

func handleMutate(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	}

	mutated, err := mutate(body, true)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(mutated)
}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/mutate", handleMutate)

	s := &http.Server{
		Addr:           ":8443",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}

	log.Fatal(s.ListenAndServeTLS("/ssl/elastic-apm-java-injector.pem", "/ssl/elastic-apm-java-injector.key"))
}
