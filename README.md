Simple utility for running arbitrary commands when the Ethereum gas price drops below a target threshold

## Installation

    go install

## Usage

    Usage: sublimate [flags] cmd
      -gwei float
    	  gas price threshold in gwei to execute command at (default 1)
      -interval int
    	  interval between price checks in seconds (default 60)
      -rpc string
    	  ethereum rpc URL

    e.g.
    
    // publish a tx once gas is below target
    sublimate cast publish 0x1234567890...

    // arbitrary commands with pipes
    sublimate -gwei 20 echo "ether.fi is pretty fly" | xargs -n1 | wc -l 
