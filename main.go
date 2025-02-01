package main

import (
	"encoding/json"
	"fmt"
	"io"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
)

func validePod(w http.ResponseWriter, r *http.Request) {
	var admissionReviewReq admissionv1.AdmissionReview

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("error could read pod object body:  \n", err)
	}
	if err := json.Unmarshal(body, &admissionReviewReq); err != nil {
		log.Printf("error could not parse pod object:  \n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	var pod corev1.Pod
	if err := json.Unmarshal(admissionReviewReq.Request.Object.Raw, &pod); err != nil {
		http.Error(w, "could not parse pod object", http.StatusBadRequest)
		log.Printf("error could not parse pod object  labels:  \n", err)
		return
	}
	labels := pod.GetLabels()
	ns := pod.GetNamespace()

	admissionReviewResp := admissionv1.AdmissionReview{
		TypeMeta: v1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID: admissionReviewReq.Request.UID,
		},
	}
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

func main() {
	http.HandleFunc("/", validePod)
	log.Println(http.ListenAndServeTLS(":443", "./tls.crt", "./tls.key", nil))
}
