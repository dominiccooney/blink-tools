// Machine larnin' on the Chromium Issue tracker.

package main

import (
	"fmt"
	"issues"
)

func main() {
	issues, err := issues.ParseIssues()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d issues, from %d to %d\n", len(issues), issues[0].Id, issues[len(issues)-1].Id)
}
