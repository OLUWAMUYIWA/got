package pkg

import (
	"errors"
	"fmt"
	"os"

	"github.com/OLUWAMUYIWA/got/pkg/proto"
)

//TODO: In doing this work, I have tried to do my best with error handling
/// Errors that do not belong directly to the Git struct
//Preference is to return these errors early after adding a context.
//the caller then decides what to do. Because this app is not robust, we mostly just panic
type OpErr struct {
	Context string
	inner   error
}

var (
	IOWriteErr    = &OpErr{Context: "Could not write to the specified writer: "}
	PermissionErr = &OpErr{Context: "Permission denied: "}
	FormatErr     = &OpErr{Context: "Bad Formatting: "}
	CopyErr       = &OpErr{Context: "Could not copy data: "}
	NotDefinedErr = &OpErr{Context: "Value not defined"}
	IOCreateErr   = &OpErr{Context: "Could not create file/ directory:"}
	IoReadErr     = &OpErr{Context: "Could not read file:"}
	Incomplete    = &OpErr{Context: "Object Incomplete"}
)

func ArgsIncomplete() error {
	argsIncompleteErr := &OpErr{Context: "Argument not complete"}
	return argsIncompleteErr

}

func (e *OpErr) Unwrap() error {
	return e.inner
}

func (e *OpErr) Error() string {
	return fmt.Sprintf("Got Op Error: %s\nInner: %v", e.Context, e.inner)
}

func (e OpErr) AddContext(s string) OpErr {
	newErr := e
	newErr.Context = fmt.Sprintf("%s: %s", newErr.Context, s)
	return newErr
}

func (e OpErr) Wrap(err error) OpErr {
	e.inner = err
	return e
}

//comback remove os.Exit(). program should only exit at main.main. here, I want to return an error value that directly causes the program to exit at the main function with a non-zero code
/// |||The Got panic error handler |||| ///
//FatalErr is a convenience function for errors that will cause the program to exit
func (git *Got) FatalErr(err error) {
	if err != nil {
		git.logger.Fatalf("got err: %v", err)
		os.Exit(1)
	}
}

//UpErr prints to the logger before returning
func (git *Got) UpErr(err error) error {
	if err != nil {
		if errors.As(err, &IOWriteErr) { //if it is of type OpErr
			git.logger.Printf("Operation:\n %v", err)
			return err
		} else if errors.As(err, &proto.GenericNetErr) {
			git.logger.Printf("Protocol:\n%v", err)
			return err
		} else {
			git.logger.Printf("Custom:\n%v", err)
			return err
		}

	}
	return nil
}
