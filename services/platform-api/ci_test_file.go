package main

// This file intentionally has a Go vet error to test CI enforcement
func UnusedFunction() {
    var x int  // unused variable - will fail go vet
}

