.PHONY: init

# docker settings
REGISTRY := "702835727665.dkr.ecr.us-east-1.amazonaws.com"
SERVICE := "polyhedron"
IMAGE := "$(REGISTRY)/$(SERVICE)"

# Docker Image management
build: 
	docker build --pull -t $(IMAGE):latest .

push: 
	docker push $(IMAGE):latest

### TESTING
local: 
	docker run -p 8090:8090 --rm $(IMAGE):latest

bash: 
	docker run --rm -it $(IMAGE):latest bash  
