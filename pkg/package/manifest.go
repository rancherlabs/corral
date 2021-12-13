package _package

type Command struct {
	NodePoolNames []string `yaml:"node_pools"`
	Command       string   `yaml:"command"`
}

type Manifest struct {
	Name        string    `yaml:"name"`
	Version     string    `yaml:"version"`
	Description string    `yaml:"description"`
	Commands    []Command `yaml:"commands"`
}
