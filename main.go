package main

import "github.com/chris-hamper/git-webhook-workflows/pkg/workflows"

func main() {
	wf := workflows.CreateWorkflow("hello-world", "workflow-template-whalesay-template", map[string]string{"message": "Hello world!"})
	workflows.WatchWorkflow(wf.Name)
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
