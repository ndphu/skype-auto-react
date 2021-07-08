package rule

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

const AllThread = "AllThread"
const AllUser = "AllUser"

type ReactRule struct {
	From   string   `json:"from"`
	To     string   `json:"to"`
	Reacts []string `json:"reacts"`
}

var configPath string

const ruleFileName = "rules.json"

var rules []ReactRule

func init() {
	dir, err := os.Getwd()
	if err == nil {
		configPath = path.Join(dir, ruleFileName)
	} else {
		configPath = ruleFileName
	}
}

func LoadRules() (error) {
	if payload, err := ioutil.ReadFile(configPath); err != nil {
		return err
	} else {
		if err := json.Unmarshal(payload, &rules); err != nil {
			return err
		}
		return nil
	}
}

func GetRules(from string, to string) []ReactRule {
	matchedRules := make([]ReactRule, 0)
	for _, rule := range rules {
		if (rule.From == AllUser || rule.From == from) && (rule.To == AllThread || rule.To == to) {
			matchedRules = append(matchedRules, rule)
		}
	}

	return matchedRules
}
