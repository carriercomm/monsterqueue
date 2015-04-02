// Copyright 2015 monsterqueue authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mongodb

import (
	"errors"
	"time"

	"github.com/tsuru/monsterqueue"
	"gopkg.in/mgo.v2/bson"
)

type jobResultMessage struct {
	Error     string
	Result    monsterqueue.JobResult
	Done      bool
	Timestamp time.Time
}

type jobOwnership struct {
	Owned     bool
	Name      string
	Timestamp time.Time
}

type JobMongoDB struct {
	Id            bson.ObjectId `bson:"_id"`
	Task          string
	Params        monsterqueue.JobParams
	Timestamp     time.Time
	Owner         jobOwnership
	ResultMessage jobResultMessage
	Waited        bool
	queue         *QueueMongoDB
}

func (j *JobMongoDB) ID() string {
	return j.Id.Hex()
}

func (j *JobMongoDB) Parameters() monsterqueue.JobParams {
	return j.Params
}

func (j *JobMongoDB) TaskName() string {
	return j.Task
}

func (j *JobMongoDB) Success(result monsterqueue.JobResult) (bool, error) {
	err := j.queue.moveToResult(j, result, nil)
	if err != nil {
		return false, err
	}
	received, err := j.queue.publishResult(j)
	return received, err
}

func (j *JobMongoDB) Error(jobErr error) (bool, error) {
	err := j.queue.moveToResult(j, nil, jobErr)
	if err != nil {
		return false, err
	}
	received, err := j.queue.publishResult(j)
	return received, err
}

func (j *JobMongoDB) Result() (monsterqueue.JobResult, error) {
	if !j.ResultMessage.Done {
		return nil, monsterqueue.ErrNoJobResult
	}
	var err error
	if j.ResultMessage.Error != "" {
		err = errors.New(j.ResultMessage.Error)
	}
	return j.ResultMessage.Result, err
}