package worker

import "fmt"

// RetryableError can be returned by a JobFunc to instruct the worker to attempt
// retry. Returning a retryable error does not guarantee that the job will be
// retried since the job would have a limit on number of retries.
type RetryableError struct {
	Cause error
}

func (re *RetryableError) Error() string {
	return fmt.Sprintf("retryable-error: %v", re.Cause)
}

func (re *RetryableError) Unwrap() error { return re.Cause }
