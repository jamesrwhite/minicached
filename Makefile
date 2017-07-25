test: build
	nohup memcached &> /dev/null &
	nohup ./minicached &> /dev/null &
	composer-vendor/bin/phpunit -v --debug --colors tests/acceptance.php
	killall memcached
	killall minicached

build:
	go build

run: build
	./minicached

connect:
	telnet localhost 5268

install:
	go install
