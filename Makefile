example-notification:
	docker-compose up --build example-notification

run-test:
	docker-compose up --build test && docker-compose stop redis