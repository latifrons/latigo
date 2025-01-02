package safe_viper

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"time"
)

func ViperMustGetString(key string) string {
	if !viper.IsSet(key) || viper.GetString(key) == "" {
		log.Panic().Str("key", key).Msg("config missing")
	}
	return viper.GetString(key)
}
func ViperMustGetInt(key string) int {
	if !viper.IsSet(key) || viper.GetString(key) == "" {
		log.Panic().Str("key", key).Msg("config missing")
	}
	return viper.GetInt(key)
}
func ViperMustGetFloat64(key string) float64 {
	if !viper.IsSet(key) || viper.GetString(key) == "" {
		log.Panic().Str("key", key).Msg("config missing")
	}
	return viper.GetFloat64(key)
}

func ViperMustGetBool(key string) bool {
	if !viper.IsSet(key) || viper.GetString(key) == "" {
		log.Panic().Str("key", key).Msg("config missing")
	}
	return viper.GetBool(key)
}

func ViperMustGetDuration(key string) time.Duration {
	if !viper.IsSet(key) || viper.GetString(key) == "" {
		log.Panic().Str("key", key).Msg("config missing")
	}
	return viper.GetDuration(key)
}
