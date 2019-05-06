package main

import (
	"sync"
	"github.com/pantheon-systems/riker/cmd"
	"github.com/pantheon-systems/riker/health"
)

// Utility function to parallelize function calls
// Useful for multiple listeners
func parallelize(functions ...func()) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(functions))

	defer waitGroup.Wait()

	for _, function := range functions {
			go func(copy func()) {
					defer waitGroup.Done()
					copy()
			}(function)
	}
}

func main() {
	parallelize(cmd.Execute, health.Start)
}
