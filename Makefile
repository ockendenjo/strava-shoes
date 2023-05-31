clean:
	find lib/ -name "*.d.ts" -delete
	find lib/ -name "*.js" -delete
	find test/ -name "*.d.ts" -delete
	find test/ -name "*.js" -delete

format:
	npx prettier --write .

deploy:
	cdk deploy --parameters "clientId=75750" ""
