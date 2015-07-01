package main

// This is the functional testing suite,
// it goes through serveral complete usage scenarios

import (
	"testing"
)

// executes following steps:
// - download and unzip to PATH
// - install daemon
// - write configuration file to repo
// - create git repo and init glass
// - wait for a timeout
// - run status and assert measured time
// - wakup with a file edit
// - run status again and assert increased time

func TestFirstTimeUse(t *testing.T) {

	//@todo implement

}
