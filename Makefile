default: docker-test

ifeq ($(DOCKER_HOST),)
    DOCKER_HOST_IP := 127.0.0.1
else
    DOCKER_HOST_IP := `echo $(DOCKER_HOST) | grep -o '[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}'`
endif

docker-test:
	EDS_ES_URL="http://$(DOCKER_HOST_IP):9200" go test

test-elasticsearch:
	docker run -d -p 9200:9200 --name elasticorm_test elasticsearch:5.2-alpine elasticsearch -Ehttp.publish_host="$(DOCKER_HOST_IP)" -Ehttp.publish_port="9200" -Ehttp.port="9200"
