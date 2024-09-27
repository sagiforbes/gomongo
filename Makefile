DOCKER_NAME = mongodb-gomongo-test
.PHONY: start-mongo
start-mongo:
	docker ps |grep $(DOCKER_NAME) || docker run -d --rm --name $(DOCKER_NAME) -p 27017:27017 mongo

.PHONY: stop-mongo
stop-mongo:
	docker ps |grep $(DOCKER_NAME) && docker stop $(DOCKER_NAME)
		
.PHONY: test-v
test-v: start-mongo
	go test -v 
.PHONY: test
test: start-mongo
	go test

.PHONY: bench
bench: start-mongo
	go test -bench=.
	
.PHONY: package
package:
	go mod tidy
	go mod vendor
