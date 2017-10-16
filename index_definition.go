package elasticorm

import (
	"errors"
	"fmt"
)

// IndexDefinition is a struct which marshals to a valid JSON configuration for creating a new elasticsearch index
type IndexDefinition struct {
	Settings IndexSettings            `json:"settings,omitempty"`
	Mappings map[string]MappingConfig `json:"mappings,omitempty"`
}

func (def IndexDefinition) elasticFieldName(typeName string, fieldName string) (string, error) {
	typeMapping, ok := def.Mappings[typeName]
	if !ok {
		return ``, errors.New(`No mapping for this type in this index definition`)
	}
	return typeMapping.elasticFieldName(fieldName)
}

type IndexSettings struct {
	NumberOfShards   int            `json:"number_of_shards,omitempty"`
	NumberOfReplicas int            `json:"number_of_replicas,omitempty"`
	Analysis         *IndexAnalysis `json:"analysis,omitempty"`
}

type IndexAnalysis struct {
	Analyzer  map[string]Analyzer  `json:"analyzer,omitempty"`
	Tokenizer map[string]Tokenizer `json:"tokenizer,omitempty"`
}

type Analyzer struct {
	Type       string   `json:"type"`
	Tokenizer  string   `json:"tokenizer"`
	CharFilter []string `json:"char_filter,omitempty"`
	Filter     []string `json:"filter,omitempty"`
}

func (d *IndexDefinition) AddAnalyzer(name string, a Analyzer) error {
	if d.Settings.Analysis.Analyzer == nil {
		d.Settings.Analysis.Analyzer = map[string]Analyzer{}
	}
	if _, ok := d.Settings.Analysis.Analyzer[name]; ok {
		return fmt.Errorf("analyzer \"%s\" already set", name)
	}
	d.Settings.Analysis.Analyzer[name] = a
	return nil
}

type Tokenizer struct {
	Type       string   `json:"type"`
	TokenChars []string `json:"token_chars,omitempty"`
	MinGram    int      `json:"min_gram,omitempty"`
	MaxGram    int      `json:"max_gram,omitempty"`
}

func (d *IndexDefinition) AddTokenizer(name string, t Tokenizer) error {
	if d.Settings.Analysis.Tokenizer == nil {
		d.Settings.Analysis.Tokenizer = map[string]Tokenizer{}
	}
	if _, ok := d.Settings.Analysis.Tokenizer[name]; ok {
		return fmt.Errorf("tokenizer \"%s\" already set", name)
	}
	d.Settings.Analysis.Tokenizer[name] = t
	return nil
}

// IndexDefinitionFunc is used as a parameter to set options on a new index definition (in NewIndexDefinition)
type IndexDefinitionFunc func(*IndexDefinition) error

// NewIndexDefinition returns a new IndexDefinition which is configurable via IndexDefinitionFuncs like SetNumberOfShards
func NewIndexDefinition(options ...IndexDefinitionFunc) (IndexDefinition, error) {
	def := IndexDefinition{}
	for _, opt := range options {
		err := opt(&def)
		if err != nil {
			return def, err
		}
	}
	for _, m := range def.Mappings {
		for _, analyzer := range m.Analyzers() {
			if analyzer == "case_insensitive_ref_id" {
				if def.Settings.Analysis == nil {
					def.Settings.Analysis = &IndexAnalysis{
						Analyzer: map[string]Analyzer{},
					}
				}
				def.Settings.Analysis.Analyzer[analyzer] = Analyzer{
					Type:      "custom",
					Tokenizer: "keyword",
					Filter:    []string{"lowercase"},
				}
			}
		}
	}
	return def, nil
}

// SetNumberOfShards is a IndexDefinitionFunc which can be passed to NewIndexDefinition and sets the number_of_shards setting
func SetNumberOfShards(number int) IndexDefinitionFunc {
	return func(def *IndexDefinition) error {
		def.Settings.NumberOfShards = number
		return nil
	}
}

// SetNumberOfReplicas is a IndexDefinitionFunc which can be passed to NewIndexDefinition and sets the number_of_recplicas setting
func SetNumberOfReplicas(number int) IndexDefinitionFunc {
	return func(def *IndexDefinition) error {
		def.Settings.NumberOfReplicas = number
		return nil
	}
}

// AddMappingFromStruct is a IndexDefinitionFunc which can be passed to NewIndexDefinition and sets the mapping for the new index by analysing the passed in struct. The mapping should be provide the functionality to save and retrieve structs of the same type (as passed in). The mapping definition is configurable via tags. See MappingFromStruct
func AddMappingFromStruct(name string, i interface{}) IndexDefinitionFunc {
	return func(def *IndexDefinition) error {
		def.Mappings = make(map[string]MappingConfig)
		mapping, err := MappingFromStruct(i)
		if err != nil {
			return err
		}
		def.Mappings[name] = mapping
		return nil
	}
}
