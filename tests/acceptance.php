<?php
require 'vendor/autoload.php';

class MinicacheTest extends PHPUnit_Framework_TestCase
{
    public function testAcceptance()
    {
        $minicache = new Memcached;
        $minicache->addServer('localhost', 5268);

        $memcached = new Memcached;
        $memcached->addServer('localhost', 11211);

        $this->assertEquals($minicache->flush(), $memcached->flush());
        $this->assertEquals($minicache->get('abc'), $memcached->get('abc'));
        $this->assertEquals($minicache->set('abc', 123), $memcached->set('abc', 123));
        $this->assertEquals($minicache->get('a'), $memcached->get('a'));
        $this->assertEquals($minicache->delete('a'), $memcached->delete('a'));
        $this->assertEquals($minicache->get('a'), $memcached->get('a'));
    }
}
