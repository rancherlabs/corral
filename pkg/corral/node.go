package corral

type Node struct {
	Name           string `yaml:"name,omitempty"`
	User           string `yaml:"user,omitempty"`
	Address        string `yaml:"address,omitempty"`
	BastionAddress string `yaml:"bastion_address,omitempty"`
}
