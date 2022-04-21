package messenger

type DeployApplication struct {
	// ID application ID
	ID uint
	// Attempt number of attempts to deploy this application
	Attempt uint
	// Commit the deployed commit
	Commit *string
	// Version the deployed version
	Version *string
}
