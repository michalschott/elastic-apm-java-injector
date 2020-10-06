package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	v1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	m "github.com/michalschott/elastic-apm-java-injector/pkg/mutate"
	log "github.com/sirupsen/logrus"
)

func mutate(body []byte) ([]byte, error) {
	log.Info("recv: %s\n", string(body))

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

		log.Debug("path: %v\n", p)
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

	log.Info("resp: %s\n", string(responseBody))

	return responseBody, nil
}

func handleMutate(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	mutated, err := mutate(body)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(mutated)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		log.Info("LOG_LEVEL not set, defaulting to info")
		logLevel = "info"
	} else {
		log.Info("LOG_LEVEL set to ", logLevel)
	}

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	switch logLevel {
	case "debug":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

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
