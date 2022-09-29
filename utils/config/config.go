package config

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/haierspi/pt-gateway/utils/ptjson"
	"github.com/robfig/config"
)

var (
	settingMutex sync.Mutex
	settingMap   = map[string]*config.Config{}
)

// Int64 read Int64
func Int64(configFile, section string, option string) int64 {
	setting := getSetting(configFile)
	s, _ := setting.String(section, option)

	if s == "" {
		log.Fatal("Configuration do not allow ", section, " ", option, " empty! ")
	} else {
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			log.Fatal("Invalid configuration:", section, " ", option, s, "! Expect an int value!")
		} else {
			return val
		}
	}
	return 0
}

// Bool read Bool
func Bool(configFile, section string, option string) bool {
	setting := getSetting(configFile)
	val, _ := setting.String(section, option)
	switch strings.ToLower(val) {
	case "true":
		return true
	case "false":
		return false
	case "":
		log.Fatal("Configuration do not allow ", section, " ", option, " empty! ")
		return false
	default:
		log.Fatal("Invalid bool configuration ", section, " ", option, " ", val)
		return false
	}
}

// String read String
func String(configFile, section string, option string) string {
	setting := getSetting(configFile)
	var val string
	val, _ = setting.String(section, option)
	if val == "" {
		log.Fatal("Configuration do not allow ", section, " ", option, " empty!")
		return ""
	}
	return val
}

// StringSlice read StringSlice
func StringSlice(configFile, section string, option string) []string {
	setting := getSetting(configFile)
	conf, _ := setting.String(section, option)
	confs := strings.Split(conf, ",")
	if len(confs) == 0 {
		log.Fatal("Configuration do not allow ", section, " ", option, " empty!")
		return nil
	}
	return confs
}

// JSON read JSON
func JSON(configFile, section string, option string) []byte {
	setting := getSetting(configFile)
	conf, _ := setting.String(section, option)

	if conf == "" {
		log.Fatal("Configuration do not allow ", section, " ", option, " empty!")
		return nil //never arrive here
	}
	if !ptjson.IsValidJSON([]byte(conf)) {
		log.Fatal("Invalid json configuration: ", conf, " in section ", section)
		return nil //never arrive here
	}
	return []byte(conf)
}

// InitConfig 初始化config
func getSetting(configFile string) *config.Config {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	if setting, ok := settingMap[configFile]; ok && setting != nil {
		return setting
	}
	if setting, err := config.ReadDefault(configFile); err == nil {
		settingMap[configFile] = setting
		return setting
	}
	log.Fatal("cannot find config file: ", configFile, " Existing...")
	return nil
}
