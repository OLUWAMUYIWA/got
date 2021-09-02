package internal

import (
	"fmt"
	"os"
)
//TODO: In doing this work, I have tried to do my best with error handling
/// Errors that do not belong directly to the Git struct
//Preference is to return these errors early after adding a context.
//the caller then decides what to do. Because this app is not robust, we mostly just panic
type OpErr struct {
	ErrSTring string
}

func (e *OpErr) Error() string {
	return e.ErrSTring
}

var (
	IOWriteErr       = &OpErr{"Could not write to the specified writer: "}
	PermissionErr = &OpErr{"Permission denied: "}
	FormatErr     = &OpErr{"Bad Formatting: "}
	OpenErr = &OpErr{"Could not open file: "}
	CopyErr = &OpErr{"COuld not copy data: "}
	NotDefinedErr = &OpErr{"Value not defined" }
	IOCreateErr = &OpErr{"Could not create file/ directory:"}
	IoReadErr = &OpErr{"Could not read file:"}
	NetworkErr = &OpErr{"Network Error: "}
)

func (e *OpErr)addContext(s string) *OpErr {
	newErr := *e
	newErr.ErrSTring = fmt.Sprintf("%s: %s",newErr.ErrSTring, s)
	return &newErr
}

/// |||The Got struct error handler |||| ///
//GotErr is a convenience function for errors that will cause the program to exit
func (git *Got)GotErr(msg interface{}) {
	if msg != nil {
		git.logger.Fatalf("got err: %v", msg)
		os.Exit(1)
	}
}
