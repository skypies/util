package tasks

import(
	"fmt"
	"net/url"
	"time"

	"golang.org/x/net/context"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// Create a new Cloud Tasks client instance. Should do this once per submission run,
// not once per task submission
// See https://godoc.org/cloud.google.com/go/cloudtasks/apiv2
func GetClient(ctxIn context.Context) (*cloudtasks.Client, error) {
	ctx,_ := context.WithTimeout(ctxIn, 30 * time.Second) // This is max deadline for cloudtasks API

	if client, err := cloudtasks.NewClient(ctx); err != nil {
		return nil, fmt.Errorf("cloudtasks.NewClient: %v", err)
	} else {
		return client, nil
	}
}

// SubmitAETask submits a new task to your App Engine queue.
func SubmitAETask(ctxIn context.Context, client *cloudtasks.Client, projectID, locationID, queueID string, wait time.Duration, uri string, params url.Values) (*taskspb.Task, error) {
	// Build the Task queue path.
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", projectID, locationID, queueID)
	
	// Build the Task payload.
	// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#CreateTaskRequest
	req := &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &taskspb.Task{			
			// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#AppEngineHttpRequest
			MessageType: &taskspb.Task_AppEngineHttpRequest{
				AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
					// For some reason, POST & body params aren't picked up by the inbound task handler
					HttpMethod:  taskspb.HttpMethod_GET,
					RelativeUri: uri + "?" + params.Encode(),
				},
			},
		},
	}

	// Apply a time delay, if needed
	if wait > 0 {
		when := timestamp.Timestamp{
			Seconds: time.Now().Add(wait).Unix(),
		}
		req.Task.ScheduleTime = &when
	}
	
	// This needs the context, even though the client already had it - bah
	ctx,_ := context.WithTimeout(ctxIn, 30 * time.Second) // This is max deadline for cloudtasks API
	createdTask, err := client.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("cloudtasks.CreateTask: %v", err)
	}

	return createdTask, nil
}
