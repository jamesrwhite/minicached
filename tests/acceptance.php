<?php
require 'vendor/autoload.php';

class MinicachedTest extends PHPUnit_Framework_TestCase
{
    public function testAcceptance()
    {
        $minicached = new Memcached;
        $minicached->addServer('localhost', 5268);

        $memcached = new Memcached;
        $memcached->addServer('localhost', 11211);

        $this->assertEquals($minicached->flush(), $memcached->flush());
        $this->assertEquals($minicached->get('abc'), $memcached->get('abc'));
        $this->assertEquals($minicached->set('abc', 123), $memcached->set('abc', 123));
        $this->assertEquals($minicached->get('a'), $memcached->get('a'));
        $this->assertEquals($minicached->delete('a'), $memcached->delete('a'));
        $this->assertEquals($minicached->get('a'), $memcached->get('a'));
    }
}
