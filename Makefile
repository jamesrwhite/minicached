test:
	go fmt
	go install
	nohup memcached &> /dev/null &
	nohup minicache &> /dev/null &
	vendor/bin/phpunit -v --debug --colors tests/acceptance.php
	killall memcached
	killall minicache

build:
	go fmt
	go build

install:
	go fmt
	go install
