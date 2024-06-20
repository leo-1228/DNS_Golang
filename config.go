package dnscheck

import (
	"bufio"
	"dnscheck/dbs"
	"dnscheck/logger"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

var Cfg *Config

type Config struct {
	PureRecords            interface{}                  `yaml:"records"`
	Workspace              string                       `yaml:"workspace"`
	BatchSize              int                          `yaml:"batch_size"`
	PreloadSize            int                          `yaml:"preload_size"`
	DomainsFile            string                       `yaml:"domains_file"`
	Spaces                 []string                     `yaml:"spaces"`
	DictionaryFile         string                       `yaml:"dictionary_file"`
	Actions                []string                     `yaml:"actions"`
	HomoglyphDouble        bool                         `yaml:"homoglyph_double"`
	SaveInvalidDomains     bool                         `yaml:"save_invalid_domains"`
	Concurrency            int                          `yaml:"concurrency"`
	PureLogMode            interface{}                  `yaml:"log_mode"`
	DNSes                  []string                     `yaml:"dnses"`
	DNS                    string                       `yaml:"dns"`
	Url                    string                       `yaml:"url"`
	ApiUrl                 string                       `yaml:"api_url"`
	CheckDuplication       bool                         `yaml:"check_duplication"`
	DuplicateDetectionMode string                       `yaml:"duplicate_detection_mode"` // "in-memory|sqlite"
	CheckMainDomains       bool                         `yaml:"check_main_domains"`
	JwtSecret              string                       `yaml:"jwt_secret"`
	Clients                map[string]*dbs.ClientConfig `yaml:"clients"`
	MainDBType             string                       `yaml:"main_db"`
	MainDBSettings         map[string]string            `yaml:"main_db_settings"`
	VariantDBType          string                       `yaml:"variant_db"`
	VariantDBSettings      map[string]string            `yaml:"variant_db_settings"`
}

func (c Config) LogMode() int {
	switch c.PureLogMode.(type) {
	case string:
		if c.PureLogMode.(string) == "realtime" {
			return 0
		}
		if c.PureLogMode.(string) == "none" {
			return -1
		}
		return 0

	case int:
		n := c.PureLogMode.(int)
		if n < -1 {
			return -1
		}
		return n
	default:
		return -1
	}
}

func (c Config) Homoglyph() bool {
	for _, v := range c.Actions {
		if v == "homoglyph" {
			return true
		}
	}
	return false
}
func (c Config) Addition() bool {
	for _, v := range c.Actions {
		if v == "addition" {
			return true
		}
	}
	return false
}
func (c Config) Bitsquatting() bool {
	for _, v := range c.Actions {
		if v == "bitsquatting" {
			return true
		}
	}
	return false
}
func (c Config) Hyphenation() bool {
	for _, v := range c.Actions {
		if v == "hyphenation" {
			return true
		}
	}
	return false
}
func (c Config) Insertion() bool {
	for _, v := range c.Actions {
		if v == "insertion" {
			return true
		}
	}
	return false
}
func (c Config) Omission() bool {
	for _, v := range c.Actions {
		if v == "omission" {
			return true
		}
	}
	return false
}
func (c Config) Repetition() bool {
	for _, v := range c.Actions {
		if v == "repetition" {
			return true
		}
	}
	return false
}
func (c Config) Replacement() bool {
	for _, v := range c.Actions {
		if v == "replacement" {
			return true
		}
	}
	return false
}
func (c Config) Transposition() bool {
	for _, v := range c.Actions {
		if v == "transposition" {
			return true
		}
	}
	return false
}
func (c Config) VowelSwap() bool {
	for _, v := range c.Actions {
		if v == "vowelSwap" {
			return true
		}
	}
	return false
}
func (c Config) Dictionary() bool {
	for _, v := range c.Actions {
		if v == "dictionary" {
			return true
		}
	}
	return false
}

func (c Config) LoadDictionary() ([]string, error) {
	if !c.Dictionary() {
		return nil, nil
	}
	file, err := os.Open(path.Join(c.Workspace, c.DictionaryFile))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make([]string, 0)
	// Create a scanner
	scanner := bufio.NewScanner(file)

	// Scan through the file line by line
	for scanner.Scan() {
		// Do something with the line
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			result = append(result, line)
		}
	}

	// Check for errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil

}
func (c Config) LoadMainDomains() ([]string, error) {
	if !c.CheckDuplication {
		return nil, nil
	}
	file, err := os.Open(path.Join(c.Workspace, c.DomainsFile))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make([]string, 0)
	// Create a scanner
	scanner := bufio.NewScanner(file)

	// Scan through the file line by line
	for scanner.Scan() {
		// Do something with the line
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			result = append(result, strings.ToLower(line))
		}
	}

	// Check for errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil

}

func (c Config) Records() (int, int, error) {

	switch c.PureRecords.(type) {
	case string:
		if c.PureRecords.(string) == "all" {
			return 0, -1, nil
		}
		return 0, 0, errors.New(fmt.Sprintf("config: records has received invalid value, %v", c.PureRecords.(string)))
	case []interface{}:
		floatSlice := make([]int, len(c.PureRecords.([]interface{})))
		for i, v := range c.PureRecords.([]interface{}) {
			floatSlice[i] = v.(int)
		}
		if len(floatSlice) == 2 {
			if floatSlice[0] >= 0 && floatSlice[1] > floatSlice[0] {
				return floatSlice[0], floatSlice[1], nil
			}
		}
		return 0, 0, errors.New(fmt.Sprintf("config: records has received invalid value, %#v", c.PureRecords.([]int)))
	default:
		return 0, 0, errors.New(fmt.Sprintf("config: records has received invalid value, %#v", c.PureRecords))
	}
}

func LoadConfig() (*Config, error) {
	yamlFile, err := os.ReadFile("./config.yaml")
	if err != nil {
		panic(err)
	}

	// Unmarshal the YAML data into a Config struct
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}

	_, _, err = config.Records()
	if err != nil {
		return &Config{}, err
	}
	if config.Concurrency <= 0 {
		config.Concurrency = 1
	}

	if config.MainDBType == "" && config.MainDBSettings == nil {
		config.MainDBType = "sqlite"
		config.MainDBSettings = make(map[string]string)
		config.MainDBSettings["path"] = "domaindata.sqlite3"
	}

	if config.VariantDBType == "" && config.VariantDBSettings == nil {
		config.VariantDBType = "sqlite"
		config.VariantDBSettings = make(map[string]string)
		config.VariantDBSettings["path"] = "variants.sqlite3"
	}

	go watchForConfigChanges(&config)

	return &config, nil
}

func watchForConfigChanges(cfg *Config) {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// Add a path.
	err = watcher.Add("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			logger.Info("event:", event)

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Remove) {
				conf2, err := LoadConfig()
				logger.Info("Got new config, applying clients configuration")
				if err != nil {
					logger.Error("Failed to read new config.yaml:", err)
				}

				cfg.Actions = conf2.Actions
				cfg.BatchSize = conf2.BatchSize
				cfg.Concurrency = conf2.Concurrency
				cfg.DNSes = conf2.DNSes
				cfg.Clients = conf2.Clients
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}

	defer watcher.Close()
}
