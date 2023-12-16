clean:
	find lib/ -name "*.d.ts" -delete
	find lib/ -name "*.js" -delete
	find test/ -name "*.d.ts" -delete
	find test/ -name "*.js" -delete

format:
	npx prettier --write .

deploy: build
	cdk deploy --stackName "StravaShoesStack" --parameters "clientId=75750"

build:
	echo "build"

synth: build
	cdk synth
