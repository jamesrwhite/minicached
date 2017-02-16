test: build
	nohup memcached &> /dev/null &
	nohup ./minicached &> /dev/null &
	vendor/bin/phpunit -v --debug --colors tests/acceptance.php
	killall memcached
	killall minicached

build:
	go build

run: build
	./minicached

install:
	go install
