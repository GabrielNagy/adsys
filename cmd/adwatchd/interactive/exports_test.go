package interactive

// InitialModelForTests returns an instance of the initial model that will not
// install the service.
func InitialModelForTests(isDefaultConfig bool) model {
	m := initialModel("adwatchd.yml", isDefaultConfig)
	m.dryrun = true
	return m
}

type Model = model
type AppConfig = appConfig
