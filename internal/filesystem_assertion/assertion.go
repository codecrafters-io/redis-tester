package filesystem_assertion

// FileSystemAssertionResult is the outcome of a single filesystem assertion run.
type FileSystemAssertionResult struct {
	SuccessLog string
	Err        error
}

type FilesystemAssertion interface {
	Run() FileSystemAssertionResult
}
