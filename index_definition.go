package elasticorm

// IndexDefinition is a struct wich marshales to a valid JSON configuration for creating a new elasticsearch index
type IndexDefinition struct {
	Settings *indexSettings         `json:"settings,omitempty"`
	Mappings map[string]interface{} `json:"mappings,omitempty"`
}

type indexSettings struct {
	NumberOfShards   int `json:"number_of_shards,omitempty"`
	NumberOfReplicas int `json:"number_of_replicas,omitempty"`
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
	return def, nil
}

// SetNumberOfShards is a IndexDefinitionFunc which can be passed to NewIndexDefinition and sets the number_of_shards setting
func SetNumberOfShards(number int) IndexDefinitionFunc {
	return func(def *IndexDefinition) error {
		if def.Settings == nil {
			def.Settings = &indexSettings{}
		}
		def.Settings.NumberOfShards = number
		return nil
	}
}

// SetNumberOfReplicas is a IndexDefinitionFunc which can be passed to NewIndexDefinition and sets the number_of_recplicas setting
func SetNumberOfReplicas(number int) IndexDefinitionFunc {
	return func(def *IndexDefinition) error {
		if def.Settings == nil {
			def.Settings = &indexSettings{}
		}
		def.Settings.NumberOfReplicas = number
		return nil
	}
}

// AddMappingFromStruct is a IndexDefinitionFunc which can be passed to NewIndexDefinition and sets the mapping for the new index by analysing the passed in struct. The mapping should be provide the functionality to save and retrieve structs of the same type (as passed in). The mapping definition is configurable via tags. See MappingFromStruct
func AddMappingFromStruct(name string, input interface{}) IndexDefinitionFunc {
	return func(def *IndexDefinition) error {
		def.Mappings = make(map[string]interface{})
		mapping, err := MappingFromStruct(input)
		if err != nil {
			return err
		}
		def.Mappings[name] = mapping
		return nil
	}
}
