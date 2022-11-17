package node

import sm "github.com/tendermint/tendermint/state"

func (n *Node) State() *sm.Store {
	return &n.stateStore
}
