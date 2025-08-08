package versioninfo

var (
	Version   = "dev"
	BuildDate = "unknown"
)

type VersionInfo struct {
	Version   string `json:"Version"`
	BuildDate string `json:"BuildDate"`
}
