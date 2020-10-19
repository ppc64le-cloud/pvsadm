package pkg

import "time"

var Options = &options{}

type options struct {
	InstanceID   string
	APIKey       string
	Region       string
	Zone         string
	DryRun       bool
	Debug        bool
	Since        time.Duration
	Before       time.Duration
	InstanceName string
	NoPrompt     bool
	IgnoreErrors bool
	AuditFile    string
	Expr         string
}

// Options for pvsadm image command
var ImageCMDOptions = &imageCMDOptions{}

type imageCMDOptions struct {
}
