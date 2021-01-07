package posse

var (
	// commitHash contains the current Git revision. Use make to build to make sure this gets set.
	commitHash string

	// buildDate contains the date of the current build.
	buildDate string
)

// Info contains information about the current posse environment
type Info struct {
	CommitHash string
	BuildDate  string
}

// Version returns the current version as a comparable version string.
func (i Info) Version() VersionString {
	return CurrentVersion.Version()
}

// NewInfo creates a new posse Info object.
func NewInfo() Info {
	return Info{
		CommitHash: commitHash,
		BuildDate:  buildDate,
	}
}
