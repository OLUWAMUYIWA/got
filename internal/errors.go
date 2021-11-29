package internal

import (
	"errors"
	"fmt"
	"os"
)

//TODO: In doing this work, I have tried to do my best with error handling
/// Errors that do not belong directly to the Git struct
//Preference is to return these errors early after adding a context.
//the caller then decides what to do. Because this app is not robust, we mostly just panic
type OpErr struct {
	ErrSTring string
	inner error
}

var (
	IOWriteErr    = &OpErr{ErrSTring: "Could not write to the specified writer: "}
	PermissionErr = &OpErr{ErrSTring: "Permission denied: "}
	FormatErr     = &OpErr{ErrSTring: "Bad Formatting: "}
	CopyErr       = &OpErr{ErrSTring: "COuld not copy data: "}
	NotDefinedErr = &OpErr{ErrSTring: "Value not defined"}
	IOCreateErr   = &OpErr{ErrSTring: "Could not create file/ directory:"}
	IoReadErr     = &OpErr{ErrSTring: "Could not read file:"}
	Incomplete 	= &OpErr{ErrSTring: "Object Incomplete"}
)


func (e *OpErr) Unwrap() error {
	return e.inner
}

func (e *OpErr) Error() string {
	return fmt.Sprintf("Got Op Error: %s\nInner: %v", e.ErrSTring, e.inner)
}


func (e OpErr) AddContext(s string) OpErr {
	newErr := e
	newErr.ErrSTring = fmt.Sprintf("%s: %s", newErr.ErrSTring, s)
	return newErr
}

func (e OpErr) Wrap(err error) OpErr {
	e.inner = err
	return e
}


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
		if e,ok := err.(*OpErr) {
			git.logger.Printf("%v",e)
			return err
		}
		
	}
}