deps:
	# adicionar hooks para o git
	git config --local core.hooksPath git-hooks/

	# baixa dependencias go modules
	go mod download

example-notification:
	docker-compose up --build example-notification

run-test:
	docker-compose up --build test && docker-compose stop redis