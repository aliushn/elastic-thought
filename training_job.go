package elasticthought

import (
	"fmt"

	"github.com/couchbaselabs/logg"
	"github.com/tleyden/go-couch"
)

// A training job represents a "training session" of a solver against training/test data
type TrainingJob struct {
	ElasticThoughtDoc
	ProcessingState ProcessingState `json:"processing-state"`
	ProcessingLog   string          `json:"processing-log"`
	UserID          string          `json:"user-id"`
	SolverId        string          `json:"solver-id" binding:"required"`

	// had to make exported, due to https://github.com/gin-gonic/gin/pull/123
	// waiting for this to get merged into master branch, since go get
	// pulls from master branch.
	Configuration Configuration
}

// Create a new training job.  If you don't use this, you must set the
// embedded ElasticThoughtDoc Type field.
func NewTrainingJob() *TrainingJob {
	return &TrainingJob{
		ElasticThoughtDoc: ElasticThoughtDoc{Type: DOC_TYPE_TRAINING_JOB},
	}
}

// Run this job
func (j TrainingJob) Run() {

	logg.LogTo("TRAINING_JOB", "Run() called!")

	// inside the job:

	// create a work directory based on config, eg, /usr/lib/elasticthought/<job-id>
	if err := j.createWorkDirectory(); err != nil {
		errMsg := fmt.Errorf("Error creating work dir: %+v.  Err: %v", j, err)
		j.recordProcessingError(errMsg)
		return
	}

	// read prototext from cbfs, do template replacement, write to work dir

	// if any env values are cbfs urls to .tar.gz files, then

	// download and extract to work dir

}

// Codereview: de-dupe
func (j TrainingJob) recordProcessingError(err error) {
	logg.LogError(err)
	db := j.Configuration.DbConnection()
	if err := j.Failed(db, err); err != nil {
		errMsg := fmt.Errorf("Error setting dataset as failed: %v", err)
		logg.LogError(errMsg)
	}
}

func (j TrainingJob) createWorkDirectory() error {
	logg.LogTo("TRAINING_JOB", "Creating dir: %v", j.Configuration.WorkDirectory)
	return mkdir(j.Configuration.WorkDirectory)

}

// Insert into database (only call this if you know it doesn't arleady exist,
// or else you'll end up w/ unwanted dupes)
// Codereview: de-dupe
func (j TrainingJob) Insert(db couch.Database) (*TrainingJob, error) {

	id, _, err := db.Insert(j)
	if err != nil {
		err := fmt.Errorf("Error inserting training job: %+v.  Err: %v", j, err)
		return nil, err
	}

	// load dataset object from db (so we have id/rev fields)
	trainingJob := &TrainingJob{}
	err = db.Retrieve(id, trainingJob)
	if err != nil {
		err := fmt.Errorf("Error fetching training job: %v.  Err: %v", id, err)
		return nil, err
	}

	return trainingJob, nil

}

// Update the state to record that it failed
// Codereview: de-dupe
func (j TrainingJob) Failed(db couch.Database, processingErr error) error {

	j.ProcessingState = Failed
	j.ProcessingLog = fmt.Sprintf("%v", processingErr)

	// TODO: retry if 409 error
	_, err := db.Edit(j)

	if err != nil {
		return err
	}

	return nil

}
