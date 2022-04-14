package corral

type Node struct {
	Name           string `json:"name,omitempty" yaml:"name,omitempty"`
	User           string `json:"user,omitempty" yaml:"user,omitempty"`
	Address        string `json:"address,omitempty" yaml:"address,omitempty"`
	BastionAddress string `json:"bastion_address,omitempty" yaml:"bastion_address,omitempty"`
	OverlayRoot    string `json:"overlay_root" yaml:"overlay_root"`
}
