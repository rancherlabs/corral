package shell

import (
	"sync"
	"time"

	"github.com/rancherlabs/corral/pkg/corral"
	"github.com/rancherlabs/corral/pkg/vars"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Registry struct {
	reg *sync.Map
}

func NewRegistry() *Registry {
	return &Registry{
		reg: &sync.Map{},
	}
}

// GetShell will return the shell associated with the given node's address.  If the shell does not exist one will be
// created.
func (r *Registry) GetShell(n corral.Node, privateKey string, vs vars.VarSet) (*Shell, error) {
	var err error

	if sh, ok := r.reg.Load(n.Address); ok {
		return sh.(*Shell), nil
	}

	err = wait.Poll(time.Second, 2*time.Minute, func() (done bool, err error) {
		sh := &Shell{
			Node:       n,
			PrivateKey: []byte(privateKey),
			Vars:       vs,
		}

		if err = sh.Connect(); err != nil {
			sh.Close()
			return false, nil
		}

		r.reg.Store(n.Address, sh)

		return err == nil, err
	})

	if err != nil {
		return nil, err
	}

	sh, _ := r.reg.Load(n.Address)
	return sh.(*Shell), err
}

// Close all shells in the registry.
func (r *Registry) Close() {
	r.reg.Range(func(key, value any) bool {
		value.(*Shell).Close()

		return true
	})
}
