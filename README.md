[![Build status](https://secure.travis-ci.org/barnybug/go-castv2.png?branch=master)](https://secure.travis-ci.org/barnybug/go-castv2)

# go-castv2

A command line tool to control Google Chromecast devices.

## Installation

Download the latest binaries from:
https://github.com/barnybug/go-castv2/releases/latest

    $ sudo mv cast-my-platform /usr/local/bin/cast
    $ sudo chmod +x /usr/local/bin/cast

## Usage

	$ cast help

Play a media file:

	$ cast --host chromecast play http://url/file.mp3

Stop playback:

	$ cast --host chromecast stop

Set volume:

	$ cast --host chromecast volume 0.5

Close app on the Chromecast:

	$ cast --host chromecast quit

## Bug reports

Please open a github issue including cast version number `cast --version`.

## Pull requests

Pull requests are gratefully received!

- please 'gofmt' the code.

## Credits

Based on go library port by [ninjasphere](https://github.com/ninjasphere/node-castv2)
