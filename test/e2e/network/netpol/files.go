package netpol

import (
	"encoding/json"
	"io/ioutil"
	"log"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type Scenarios struct {
	Content []TestData `json:"scenarios"`
}

type TestData struct {
	Netpols []networkingv1.NetworkPolicy `json:"netpols"`

	FromList []string `json:"fromList,flow"`
	ToList   []string `json:"toList,flow"`

	Title   string    `json:"title"`
	Matches []Matches `json:"matches"`
}

type Matches struct {
	From string   `yaml:"from"`
	To   []string `yaml:"to,flow"`
}

func LoadYAMLTests() Scenarios {
	scenarios := Scenarios{}
	yamlContent, err := ioutil.ReadFile("tests/tests.yaml")
	if err != nil {
		log.Fatal("ERROR loading yaml: ", err)
	}

	// Convert to JSON so netpol items can be unmarshalled by api machinery
	jsonContent, err := yaml.ToJSON(yamlContent)
	if err != nil {
		log.Fatal("ERROR unmarshalling yaml: ", err)
	}

	err = json.Unmarshal(jsonContent, &scenarios)
	if err != nil {
		log.Fatal("ERROR unmarshalling yaml: ", err)
	}

	return scenarios
}

// BuildTruthTable create expected thruth table
func BuildTruthTable(td TestData) *TruthTable {
	value := true
	tdd := NewTruthTable(td.FromList, td.ToList, &value)
	for _, r := range td.Matches {
		for _, toItem := range td.ToList {
			if !isInsideList(toItem, r.To) {
				tdd.Set(r.From, toItem, false)
			}
		}
	}
	return tdd
}

func isInsideList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
