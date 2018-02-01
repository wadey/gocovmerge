docker-image:
	CGO_ENABLED=0 GOOS=linux go build -o build/linux/gocovmerge
	docker build -t wadey/gocovmerge .

clean:
	go clean
	$(RM) -r build
