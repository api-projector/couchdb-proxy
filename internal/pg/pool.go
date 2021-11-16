package pg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
	"log"
)

const (
	paramPgUser     = "pg_user"
	paramPgPassword = "pg_password"
	paramPgHost     = "pg_host"
	paramPgPort     = "pg_port"
	paramPgDatabase = "pg_database"
)

var pgParams = []string{paramPgUser, paramPgPassword, paramPgHost, paramPgPort, paramPgDatabase}

type pgConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
}

func (config *pgConfig) GetConnectionString() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		config.User, config.Password, config.Host, config.Port, config.Database,
	)
}

func GetConnectionPool() *pgxpool.Pool {
	config := readPgConfig()

	log.Printf(
		"create pg pool: host=%s, port=%d, user=%s, db=%s",
		config.Host, config.Port, config.User, config.Database,
	)

	poolConfig, err := pgxpool.ParseConfig(config.GetConnectionString())
	if err != nil {
		panic(err)
	}

	poolConfig.ConnConfig.PreferSimpleProtocol = true

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		panic(err)
	}

	return pool
}

func readPgConfig() *pgConfig {
	for _, env := range pgParams {
		err := viper.BindEnv(env)
		if err != nil {
			panic(err)
		}
	}

	viper.SetDefault(paramPgPort, "5432")

	return &pgConfig{
		Port:     viper.GetInt(paramPgPort),
		Host:     viper.GetString(paramPgHost),
		User:     viper.GetString(paramPgUser),
		Password: viper.GetString(paramPgPassword),
		Database: viper.GetString(paramPgDatabase),
	}
}
