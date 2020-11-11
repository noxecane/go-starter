package config

// Env is the expected config values from the process's environment
type Env struct {
	AppEnv string `default:"dev" split_words:"true"`
	Name   string `required:"true"`
	Port   int    `required:"true"`
	Scheme string `required:"true"`
	Secret []byte `required:"true"`

	TemplateDir string `required:"true" split_words:"true"`

	PostgresHost       string `required:"true" split_words:"true"`
	PostgresPort       int    `required:"true" split_words:"true"`
	PostgresSecureMode bool   `required:"true" split_words:"true"`
	PostgresUser       string `required:"true" split_words:"true"`
	PostgresPassword   string `required:"true" split_words:"true"`
	PostgresDatabase   string `required:"true" split_words:"true"`

	RedisHost     string `required:"true" split_words:"true"`
	RedisPort     int    `required:"true" split_words:"true"`
	RedisPassword string `default:"" split_words:"true"`

	SendgridKey     string `required:"true" split_words:"true"`
	MailSender      string `required:"true" split_words:"true"`
	NotifyEmail     string `required:"true" split_words:"true"`
	PostmasterEmail string `required:"true" split_words:"true"`

	SessionTimeout  string `required:"true" split_words:"true"`
	HeadlessTimeout string `required:"true" split_words:"true"`

	ClientOwnerPage string `required:"true" split_words:"true"`
	ClientUserPage  string `required:"true" split_words:"true"`
	ClientResetPage string `required:"true" split_words:"true"`
}
