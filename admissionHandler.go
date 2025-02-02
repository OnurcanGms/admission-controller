package main

import (
	"encoding/json"
	"fmt"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
)

// validePod handles pod creation requests the Kubernetes cluster environment.
// It verifies if a pod's namespace matches the "namespace" label and rejects pod creation if there is no match.
// Responds to the AdmissionReview request with an allowed/denied status and an appropriate message.
// Created to prevent deploying apps to wrong namespaces
func validePod(w http.ResponseWriter, r *http.Request) {

	var admissionReviewReq admissionv1.AdmissionReview

	// decode body from kubernetes pod creation request and if there is an error
	// log that error and return
	if err := json.NewDecoder(r.Body).Decode(&admissionReviewReq); err != nil {
		log.Printf("error could not parse pod object: %s \n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// unmarshal to coming request to pod struct for get pod information
	var pod corev1.Pod
	if err := json.Unmarshal(admissionReviewReq.Request.Object.Raw, &pod); err != nil {
		http.Error(w, "could not parse pod object", http.StatusBadRequest)
		log.Printf("error could not parse pod object  labels: %s \n", err)
		return
	}
	labels := pod.GetLabels()
	ns := pod.GetNamespace()
	// create a response base
	admissionReviewResp := admissionv1.AdmissionReview{
		TypeMeta: v1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID: admissionReviewReq.Request.UID,
		},
	}
	// check pod labels and namespaces to allow or deny pod creation request
	if val, ok := labels["namespace"]; ok && val == ns {
		admissionReviewResp.Response.Allowed = true
		admissionReviewResp.Response.Result = &v1.Status{
			Message: fmt.Sprintf("pod namespace (%s) is matching with label namespace (%s) is allowed", ns, val),
		}
	}
	if val, ok := labels["namespace"]; ok && val != ns {
		admissionReviewResp.Response.Allowed = false
		admissionReviewResp.Response.Result = &v1.Status{
			Message: fmt.Sprintf("Pod creation restricted because pod namespace (%s) does not match label namespace (%s)", ns, val),
		}
	}
	if _, ok := labels["namespace"]; !ok {
		admissionReviewResp.Response.Allowed = false
		admissionReviewResp.Response.Result = &v1.Status{
			Message: "Pod creation restricted because namespace label  does not exist",
		}
	}
	if len(labels) == 0 {
		admissionReviewResp.Response.Allowed = false

		admissionReviewResp.Response.Result = &v1.Status{
			Message: "Pod creation restricted because it has no label",
		}
	}
	// create response for kubernetes with new response object
	resp, err := json.Marshal(admissionReviewResp)
	if err != nil {
		log.Printf("error could not encode response:  %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("error could not write response:  %v\n", err)
	}
}
