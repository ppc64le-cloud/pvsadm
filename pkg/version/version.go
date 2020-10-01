package version

// set version while building the binary, otherwise unknown is the version
var Version = "unknown"

func Get() string {
	return Version
}
