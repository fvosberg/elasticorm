package elasticorm

type IndexDefinition struct {
	Settings *indexSettings `json:"settings,omitempty"`
}

type indexSettings struct {
	NumberOfShards int `json:"number_of_shards"`
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
