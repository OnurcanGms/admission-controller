package main

import (
	"bytes"
	"encoding/json"
	"io"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"net/http/httptest"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidePod(t *testing.T) {
	tests := []struct {
		name       string
		body       admissionv1.AdmissionReview
		statusCode int
		allowed    bool
		message    string
	}{
		{
			name: "Success - namespace matches label",
			body: admissionv1.AdmissionReview{
				Request: &admissionv1.AdmissionRequest{
					UID: "test-uid",
					Object: runtime.RawExtension{
						Raw: encodePod(&corev1.Pod{
							ObjectMeta: v1.ObjectMeta{
								Namespace: "test-namespace",
								Labels: map[string]string{
									"namespace": "test-namespace",
								},
							},
						}),
					},
				},
			},
			statusCode: http.StatusOK,
			allowed:    true,
			message:    "pod namespace (test-namespace) is matching with label namespace (test-namespace) is allowed",
		},
		{
			name: "Fail - namespace differs from label",
			body: admissionv1.AdmissionReview{
				Request: &admissionv1.AdmissionRequest{
					UID: "test-uid",
					Object: runtime.RawExtension{
						Raw: encodePod(&corev1.Pod{
							ObjectMeta: v1.ObjectMeta{
								Namespace: "test-namespace",
								Labels: map[string]string{
									"namespace": "different-namespace",
								},
							},
						}),
					},
				},
			},
			statusCode: http.StatusOK,
			allowed:    false,
			message:    "Pod creation restricted because pod namespace (test-namespace) does not match label namespace (different-namespace)",
		},
		{
			name: "Fail - no namespace label",
			body: admissionv1.AdmissionReview{
				Request: &admissionv1.AdmissionRequest{
					UID: "test-uid",
					Object: runtime.RawExtension{
						Raw: encodePod(&corev1.Pod{
							ObjectMeta: v1.ObjectMeta{
								Namespace: "test-namespace",
								Labels: map[string]string{
									"different-label": "different-value",
								},
							},
						}),
					},
				},
			},
			statusCode: http.StatusOK,
			allowed:    false,
			message:    "Pod creation restricted because namespace label  does not exist",
		},
		{
			name: "Fail - no labels at all",
			body: admissionv1.AdmissionReview{
				Request: &admissionv1.AdmissionRequest{
					UID: "test-uid",
					Object: runtime.RawExtension{
						Raw: encodePod(&corev1.Pod{
							ObjectMeta: v1.ObjectMeta{
								Namespace: "test-namespace",
							},
						}),
					},
				},
			},
			statusCode: http.StatusOK,
			allowed:    false,
			message:    "Pod creation restricted because it has no label",
		},
		{
			name:       "Fail - malformed request body",
			body:       admissionv1.AdmissionReview{},
			statusCode: http.StatusBadRequest,
			allowed:    false,
			message:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyData []byte
			if tt.body.Request != nil {
				bodyData, _ = json.Marshal(tt.body)
			} else {
				bodyData = []byte("invalid-body")
			}

			req := httptest.NewRequest(http.MethodPost, "/validatePod", bytes.NewReader(bodyData))
			rec := httptest.NewRecorder()

			validePod(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.statusCode {
				t.Errorf("expected status code %d, got %d", tt.statusCode, res.StatusCode)
			}

			if tt.statusCode == http.StatusOK {
				var admissionResp admissionv1.AdmissionReview
				respBody, _ := io.ReadAll(res.Body)
				json.Unmarshal(respBody, &admissionResp)

				response := admissionResp.Response
				if response.Allowed != tt.allowed {
					t.Errorf("expected allowed %v, got %v", tt.allowed, response.Allowed)
				}
				if tt.message != "" && response.Result.Message != tt.message {
					t.Errorf("expected message %q, got %q", tt.message, response.Result.Message)
				}
			}
		})
	}
}

func encodePod(pod *corev1.Pod) []byte {
	podData, _ := json.Marshal(pod)
	return podData
}
