<?php

$minicache = new Memcached;
$minicache->addServer('localhost', 5268);

var_dump($minicache->get('a'));
var_dump($minicache->set('a', array(1, 2, 3)));
var_dump($minicache->get('a'));
var_dump($minicache->delete('a'));
var_dump($minicache->set('a', 123));
var_dump($minicache->get('a'));
var_dump($minicache->flush());
