<?php

$minicache = new Memcached;
$minicache->addServer('localhost', 5268);

$memcached = new Memcached;
$memcached->addServer('localhost', 11211);

function compare($minicache, $memcached) {
	if ($minicache !== $memcached) {
		echo 'minicache' . PHP_EOL;
		var_dump($minicache);
		echo PHP_EOL;

		echo 'memcached' . PHP_EOL;
		var_dump($memcached);
		echo PHP_EOL;

		throw new Exception('Result for minicache did not match memcached!');
	}

	return true;
}

compare($minicache->flush(), $memcached->flush());
compare($minicache->get('abc'), $memcached->get('abc'));
compare($minicache->set('abc', 123), $memcached->set('abc', 123));
compare($minicache->get('a'), $memcached->get('a'));
compare($minicache->delete('a'), $memcached->delete('a'));
compare($minicache->get('a'), $memcached->get('a'));
