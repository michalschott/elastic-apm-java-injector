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

	m "github.com/michalschott/elastic-apm-java-injector/pkg/mutate"
)

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
		p := []m.PatchOperation{}

		p = append(p, m.AddVolume(pod)...)
		p = append(p, m.AddInitContainer(pod)...)
		p = append(p, m.MutateContainers(pod)...)

		resp.Patch, err = json.Marshal(p)
		if err != nil {
			return nil, err
		}

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
	_, err = w.Write(mutated)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	}
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
