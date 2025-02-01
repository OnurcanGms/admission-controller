# Build Docker image
build:
	docker image rm -f myregistry/admission-webhook:v1 || true
	docker build -t myregistry/admission-webhook:v1 ../
# Delete all Kubernetes resources
clean:
	kubectl delete -f ./admissionDeploy.yaml  --ignore-not-found=true
	kubectl delete -f ./admissionSecret.yaml  --ignore-not-found=true
	kubectl delete -f ./admission-service.yaml  --ignore-not-found=true
	kubectl delete -f ./admission-webhook.yaml  --ignore-not-found=true
#Clean and apply Kubernetes resources
deploy: build clean
	kubectl apply -f admissionDeploy.yaml
	kubectl apply -f admissionSecret.yaml
	kubectl apply -f admission-service.yaml
	kubectl apply -f admission-webhook.yaml
#Only apply kubernetes resource
apply:
	kubectl apply -f admissionDeploy.yaml
	kubectl apply -f admissionSecret.yaml
	kubectl apply -f admission-service.yaml
	kubectl apply -f admission-webhook.yaml
#Apply test resource
test:
	kubectl apply -f test.yaml