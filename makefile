cartographer_provider: $(wildcard *.go)
	go build -o terraform-provider-aws-uncontrolled

%.cartographer.zip: $(wildcard package/*.js)
	(cd package && zip ../$@ *)