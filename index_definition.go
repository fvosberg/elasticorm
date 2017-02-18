package elasticorm

type IndexDefinition struct {
	Settings *indexSettings         `json:"settings,omitempty"`
	Mappings map[string]interface{} `json:"mappings,omitempty"`
}

type indexSettings struct {
	NumberOfShards   int `json:"number_of_shards,omitempty"`
	NumberOfReplicas int `json:"number_of_replicas,omitempty"`
}

type IndexDefinitionFunc func(*IndexDefinition) error

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

func SetNumberOfShards(number int) IndexDefinitionFunc {
	return func(def *IndexDefinition) error {
		if def.Settings == nil {
			def.Settings = &indexSettings{}
		}
		def.Settings.NumberOfShards = number
		return nil
	}
}

func SetNumberOfReplicas(number int) IndexDefinitionFunc {
	return func(def *IndexDefinition) error {
		if def.Settings == nil {
			def.Settings = &indexSettings{}
		}
		def.Settings.NumberOfReplicas = number
		return nil
	}
}

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
