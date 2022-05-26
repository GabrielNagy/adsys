package interactive

// InitialModelForTests returns an instance of the initial model that will not
// install the service.
func InitialModelForTests() model {
	m := initialModel("adwatchd.yaml")
	m.dryrun = true
	return m
}

type Model = model
