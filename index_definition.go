package elasticorm

type IndexDefinition struct {
	Settings *indexSettings `json:"settings,omitempty"`
}

type indexSettings struct {
	NumberOfShards   int `json:"number_of_shards,omitempty"`
	NumberOfReplicas int `json:"number_of_replicas,omitempty"`
}

type IndexDefinitionFunc func(*IndexDefinition)

func NewIndexDefinition(options ...IndexDefinitionFunc) IndexDefinition {
	def := IndexDefinition{}
	for _, opt := range options {
		opt(&def)
	}
	return def
}

func SetNumberOfShards(number int) IndexDefinitionFunc {
	return func(def *IndexDefinition) {
		if def.Settings == nil {
			def.Settings = &indexSettings{}
		}
		def.Settings.NumberOfShards = number
	}
}

func SetNumberOfReplicas(number int) IndexDefinitionFunc {
	return func(def *IndexDefinition) {
		if def.Settings == nil {
			def.Settings = &indexSettings{}
		}
		def.Settings.NumberOfReplicas = number
	}
}
