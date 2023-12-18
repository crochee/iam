package mysql

// NewMysqlClient create mysql factory with context.Context
func NewClient(ctx context.Context) (*dataStore, error) {
	c, err := storage.New(ctx,
		storage.WithUser(viper.GetString("mysql.user")),
		storage.WithPassword(viper.GetString("mysql.password")),
		storage.WithIP(viper.GetString("mysql.ip")),
		storage.WithPort(viper.GetString("mysql.port")),
		storage.WithDatabase(viper.GetString("mysql.name")),
		storage.WithCharset(viper.GetString("mysql.charset")),
		storage.WithMaxOpenConn(viper.GetInt("mysql.max_open_conns")),
		storage.WithMaxIdleConn(viper.GetInt("mysql.max_idle_conns")),
		storage.WithMaxLifetime(time.Duration(viper.GetInt("mysql.conn_max_lifetime"))*time.Second),
		storage.WithLogger(storage.NewLog(slog.Default())),
		storage.WithPlugins(
			storage.IgnoreSelectLogger{},
		))
	if err != nil {
		return nil, err
	}
	return &dataStore{DB: c}, nil
}

type dataStore struct {
	*storage.DB
}
