package main

import (
	"io"

	"github.com/adewale/rogue_planet/pkg/logging"
)

// ErrUserCancelled indicates the user cancelled an operation
type ErrUserCancelled struct {
	message string
}

func (e *ErrUserCancelled) Error() string {
	return e.message
}

// Command options structures for testability

type InitOptions struct {
	FeedsFile  string
	ConfigPath string
	Output     io.Writer
}

type AddFeedOptions struct {
	URL        string
	ConfigPath string
	Output     io.Writer
}

type AddAllOptions struct {
	FeedsFile  string
	ConfigPath string
	Output     io.Writer
}

type RemoveFeedOptions struct {
	URL        string
	ConfigPath string
	Output     io.Writer
	Input      io.Reader // For reading confirmation (testable)
	Force      bool      // Skip confirmation prompt
}

type ListFeedsOptions struct {
	ConfigPath string
	Output     io.Writer
}

type StatusOptions struct {
	ConfigPath string
	Output     io.Writer
}

type UpdateOptions struct {
	ConfigPath string
	Verbose    bool
	Output     io.Writer
	Logger     logging.Logger
}

type FetchOptions struct {
	ConfigPath string
	Verbose    bool
	Output     io.Writer
	Logger     logging.Logger
}

type GenerateOptions struct {
	ConfigPath string
	Days       int
	Output     io.Writer
}

type PruneOptions struct {
	ConfigPath string
	Days       int
	DryRun     bool
	Output     io.Writer
}

type VerifyOptions struct {
	ConfigPath string
	Output     io.Writer
}

type ImportOPMLOptions struct {
	OPMLFile   string
	ConfigPath string
	DryRun     bool
	Output     io.Writer
}

type ExportOPMLOptions struct {
	OutputFile string
	ConfigPath string
	Output     io.Writer
}
