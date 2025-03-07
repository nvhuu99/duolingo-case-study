package workdistributor

import (
	"fmt"
	"runtime"
)

const (
	UnKnownErr					= 500
	ConnectionFailure			= 501
	LockErr						= 502
	WorkloadEmpty				= 503
	WorkloadNotFound			= 504
	WorkloadNotSet				= 505
	WorkloadAttributesInvalid	= 506
	WorkloadCreationErr			= 507
	WorkloadDuplication			= 508
	WorkloadSwitchErr			= 509
	WorkloadAssignmentErr		= 510
	PurgeFailure				= 511
)

var ErrMessages = map[int]string{
	UnKnownErr:					"unknown error",
	ConnectionFailure:			"connection failure",
	LockErr:					"can not acquire lock",
	WorkloadEmpty:				"workload empty",
	WorkloadNotFound:			"workload does not exist",
	WorkloadNotSet:				"workload not set",
	WorkloadAttributesInvalid:	"workload attributes invalid",
	WorkloadCreationErr:		"fail to register new workload",
	WorkloadDuplication:		"workload duplication",
	WorkloadSwitchErr:			"workload duplication",
	WorkloadAssignmentErr:		"workload assignment error",
	PurgeFailure:				"fail to purge workloads",
}

type Error struct {
	DistributorName	string
	WorkloadName	string
	AssignmentId	string
	
	Code          int
	Message       string
	OriginalError error
	
	FuncName   string
	File       string
	LineNumber int
}

func (e *Error) Error() string {
	mssg := fmt.Sprintf(
		"error code: %v, error message: \"%v\"\nfile: \"%v\", line: %v, function: \"%v\"\n",
		e.Code,
		e.Message,
		e.File,
		e.LineNumber,
		e.FuncName,
	)

	if e.DistributorName != "" {
		mssg += "distributor name: " + e.DistributorName
		if e.WorkloadName != "" {
			mssg += ", workload name: " + e.WorkloadName
		}
		if e.AssignmentId != "" {
			mssg += ", assignment id: " + e.AssignmentId
		}
		mssg += "\n"
	}

	if e.OriginalError != nil {
		if _, check := e.OriginalError.(*Error); check {
			return mssg +  e.OriginalError.Error()
		} else {
			mssg += fmt.Sprintf("original error: %v", e.OriginalError)
		}
	}

	return mssg
}

func NewError(code int, err error, distributor string, workload string, assignment string) *Error {
	mqErr := Error {
		Code: code,
		Message: ErrMessages[code],
		OriginalError: err,
		DistributorName: distributor,
		WorkloadName: workload,
		AssignmentId: assignment,
	}
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		mqErr.File = file
		mqErr.LineNumber = line
	}
	if f := runtime.FuncForPC(pc); f != nil {
		mqErr.FuncName = f.Name()
	}

	return &mqErr
}
