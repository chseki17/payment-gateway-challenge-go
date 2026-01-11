package config

type Config struct {
	App           AppConfig
	BankSimulator BankSimulatorConfig
}

type AppConfig struct {
	Name        string `envconfig:"APP_NAME"        default:"payment_gateway_api"`
	Environment string `envconfig:"APP_ENVIRONMENT" default:"development"`
	Version     string `envconfig:"APP_VERSION"     default:"v1"`
	LogLevel    string `envconfig:"APP_LOG_LEVEL"   default:"info"`
	APIPort     string `envconfig:"APP_API_PORT"    default:"8090"`
}

type BankSimulatorConfig struct {
	URL string `envconfig:"BANK_SIMULATOR_URL" default:"http://localhost:8080"`
}
