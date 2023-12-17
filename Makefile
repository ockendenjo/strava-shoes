.PHONY: build test synth

clean:
	rm -rf build
	rm -rf cdk.out

test:
	bash -c 'diff -u <(echo -n) <(go fmt $(go list ./...))'
	go vet ./...
	go test ./... -v && echo "\nResult=OK" || (echo "\nResult=FAIL" && exit 1)

deploy: build
	cdk deploy --stackName "StravaShoesStack" --parameters "clientId=75750"

build:
	go run scripts/build-cmd/main.go

synth: build
	cdk synth
