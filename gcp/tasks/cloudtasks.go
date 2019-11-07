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

// SubmitAETask submits a new task to your App Engine queue.
func SubmitAETask(ctxIn context.Context, projectID, locationID, queueID, wait time.Duration, uri string, params url.Values) (*taskspb.Task, error) {
	// Create a new Cloud Tasks client instance.
	// See https://godoc.org/cloud.google.com/go/cloudtasks/apiv2

	ctx,_ := context.WithTimeout(ctxIn, 30 * time.Second) // This is max deadline for cloudtasks API
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("cloudtasks.NewClient: %v", err)
	}

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
	
	createdTask, err := client.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("cloudtasks.CreateTask: %v", err)
	}

	return createdTask, nil
}
