package workflows

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	wfclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	wfv1cs "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"
)

var (
	wfClient wfv1cs.WorkflowInterface
)

func init() {
	// Try to get kubeconfig from the command line or environment
	kubeconfig := flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "(optional) absolute path to the kubeconfig file")
	flag.Parse()

	// Use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	panicOnError(err)

	// Create the workflow client
	namespace := strings.TrimSpace(os.Getenv("POD_NAMESPACE"))
	if namespace == "" {
		panic(fmt.Errorf("POD_NAMESPACE env var must be set"))
	}

	wfClient = wfclientset.NewForConfigOrDie(config).ArgoprojV1alpha1().Workflows(namespace)
}

func CreateWorkflow(name, templateName string, parameters map[string]string) *wfv1.Workflow {
	var args wfv1.Arguments
	for k, v := range parameters {
		value := wfv1.AnyString(v)
		args.Parameters = append(args.Parameters, wfv1.Parameter{Name: k, Value: &value})
	}

	var workflow = wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
		},
		Spec: wfv1.WorkflowSpec{
			Entrypoint: "entry",
			Templates: []wfv1.Template{
				{
					Name: "entry",
					Steps: []wfv1.ParallelSteps{
						{
							Steps: []wfv1.WorkflowStep{
								{
									Name: "run-handler",
									TemplateRef: &wfv1.TemplateRef{
										Name: templateName,
										Template: "whalesay-template",
										// Template: "handler",
									},
									Arguments: args,
								},
							},
						},
					},
				},
			},
		},
	}

	// Submit the workflow
	created, err := wfClient.Create(context.TODO(), &workflow, metav1.CreateOptions{})
	panicOnError(err)
	fmt.Printf("Workflow %s created\n", created.Name)
	
	return created
}

func WatchWorkflow(name string) {
	// Wait for the workflow to complete
	fieldSelector := fields.ParseSelectorOrDie(fmt.Sprintf("metadata.name=%s", name))
	watchIf, err := wfClient.Watch(context.TODO(), metav1.ListOptions{FieldSelector: fieldSelector.String(), TimeoutSeconds: pointer.Int64Ptr(180)})
	panicOnError(err)
	defer watchIf.Stop()
	for next := range watchIf.ResultChan() {
		wf, ok := next.Object.(*wfv1.Workflow)
		if !ok {
			continue
		}
		if !wf.Status.FinishedAt.IsZero() {
			fmt.Printf("Workflow %s %s at %v. Message: %s.\n", wf.Name, wf.Status.Phase, wf.Status.FinishedAt, wf.Status.Message)
			break
		}
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
