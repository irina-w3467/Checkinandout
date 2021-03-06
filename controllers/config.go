package controllers

import (
	"log"
	"path/filepath"

	svc "cloudminds.com/harix/cc-server/services"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
)

// EmailConfig - for sending Email to activate user
type EmailConfig struct {
	SMTPHost      string `json:"smtp_host" mapstructure:"smtp_host"`
	SMTPPort      string `json:"smtp_port" mapstructure:"smtp_port"`
	FromEmailAddr string `json:"from_email_addr" mapstructure:"from_email_addr"`
	FromEmailPswd string `json:"from_email_pswd" mapstructure:"from_email_pswd"`
}

// SMSConfig - for sending SMS to activate user
type SMSConfig struct {
	AccountSid   string `json:"account_sid" mapstructure:"account_sid"`
	AuthToken    string `json:"auth_token" mapstructure:"auth_token"`
	FromPhoneNum string `json:"from_phone_num" mapstructure:"from_phone_num"`
}

type DebugTokenList struct {
	SuperAdmin string `json:"super_admin" mapstructure:"super_admin"`
	Admin      string `json:"admin" mapstructure:"admin"`
	Mobile     string `json:"mobile" mapstructure:"mobile"`
}

// Config - top-level configuration structure
type Config struct {
	MongoServerURI      string         `json:"mongo_server_uri" mapstructure:"mongo_server_uri"`
	ServerAddr          string         `json:"server_address" mapstructure:"server_address"`
	RequireCheckOutTemp bool           `json:"require_check_out_temperature" mapstructure:"require_check_out_temperature"`
	RequireAdminPswd    bool           `json:"require_admin_password" mapstructure:"require_admin_password"`
	TempThrd            float32        `json:"temperature_threshold" mapstructure:"temperature_threshold"`
	DebugTokenL         DebugTokenList `json:"debug_token_list" mapstructure:"debug_token_list"`
	EmailConf           EmailConfig    `json:"email_config" mapstructure:"email_config"`
	SMSConf             SMSConfig      `json:"sms_config" mapstructure:"sms_config"`
}

// var defaulEmailConfig = EmailConfig{
// 	SMTPHost:      "smtp.gmail.com",
// 	SMTPPort:      "587",
// 	FromEmailAddr: "notifications.cc.app@gmail.com",
// 	FromEmailPswd: "Wrightrobotics123",
// }

// var defaultSMSConfig = SMSConfig{
// 	AccountSid:   "AC61389296221b860447ed00967abf77b5",
// 	AuthToken:    "",
// 	FromPhoneNum: "+19169933295",
// }

var defaultConfig = Config{
	RequireCheckOutTemp: false,
}

// InitConfig - loading global configurations from json file
func (s *CCServer) InitConfig(appName string, isTest bool) {
	if isTest {
		viper.AddConfigPath(filepath.Join("..", "configs"))
	} else {
		viper.AddConfigPath(filepath.Join(".", "configs"))
	}
	viper.SetConfigName(appName)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Viper - Error when reading config file %v\n", err)
	}
	log.Println("Using config file:", viper.ConfigFileUsed())
	// log.Printf("InitConfig - log_file_name: %v\n", viper.GetString("log_file_name"))
	config := defaultConfig
	err := viper.UnmarshalKey("services", &config)
	if err != nil {
		log.Printf("init config failed with error - %v\n", err)
	}
	// log.Printf("config initialized - %v\n", config)

	s.Config = config
	log.Printf("config initiated, tokens: %v\n", s.Config.DebugTokenL)
}

// ReloadConfigFromDB - reload hot-reloadable configs from DB
func (s *CCServer) ReloadConfigFromDB() {
	var reloadedConfig svc.Config
	err := svc.GetConfigByName("default").Decode(&reloadedConfig)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("Default Config does not exist")
			return
		}
		log.Println("Error while getting Default Config by Name")
		return
	}

	// log.Println("SMSConf.AuthToken: ", reloadedConfig.SMSAuthToken)
	s.Config.SMSConf.AuthToken = reloadedConfig.SMSAuthToken
	s.Config.ServerAddr = reloadedConfig.ServerAddr
	log.Printf("config reloaded - %v\n", s.Config)

}
