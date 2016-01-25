test:
	go fmt
	go install
	nohup memcached &> /dev/null &
	nohup minicached &> /dev/null &
	vendor/bin/phpunit -v --debug --colors tests/acceptance.php
	killall memcached
	killall minicached

build:
	go fmt
	go build

install:
	go fmt
	go install
