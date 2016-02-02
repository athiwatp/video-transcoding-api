package provider

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
)

type failure struct {
	op  string
	err error
}

type fakeElasticTranscoder struct {
	*elastictranscoder.ElasticTranscoder
	jobs     map[string]*elastictranscoder.CreateJobInput
	failures chan failure
}

func newFakeElasticTranscoder() *fakeElasticTranscoder {
	return &fakeElasticTranscoder{
		ElasticTranscoder: &elastictranscoder.ElasticTranscoder{},
		failures:          make(chan failure, 1),
		jobs:              make(map[string]*elastictranscoder.CreateJobInput),
	}
}

func (c *fakeElasticTranscoder) CreateJob(input *elastictranscoder.CreateJobInput) (*elastictranscoder.CreateJobResponse, error) {
	if err := c.getError("CreateJob"); err != nil {
		return nil, err
	}
	id := "job-" + generateID(4)
	c.jobs[id] = input
	return &elastictranscoder.CreateJobResponse{
		Job: &elastictranscoder.Job{
			Id:         aws.String(id),
			Input:      input.Input,
			PipelineId: input.PipelineId,
			Status:     aws.String("Submitted"),
		},
	}, nil
}

func (c *fakeElasticTranscoder) ReadJob(input *elastictranscoder.ReadJobInput) (*elastictranscoder.ReadJobOutput, error) {
	if err := c.getError("ReadJob"); err != nil {
		return nil, err
	}
	createJobInput, ok := c.jobs[*input.Id]
	if !ok {
		return nil, errors.New("job not found")
	}
	return &elastictranscoder.ReadJobOutput{
		Job: &elastictranscoder.Job{
			Id:         input.Id,
			Input:      createJobInput.Input,
			PipelineId: createJobInput.PipelineId,
			Status:     aws.String("Complete"),
		},
	}, nil
}

func (c *fakeElasticTranscoder) prepareFailure(op string, err error) {
	c.failures <- failure{op: op, err: err}
}

func (c *fakeElasticTranscoder) getError(op string) error {
	select {
	case prepFailure := <-c.failures:
		if prepFailure.op == op {
			return prepFailure.err
		}
		c.failures <- prepFailure
	default:
	}
	return nil
}
