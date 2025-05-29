package operator

import "github.com/reearth/reearthx/account/accountdomain"

type Operator struct {
	user        *accountdomain.UserID
	integration *IntegrationID
	isMachine   bool
}

func FromUser(user accountdomain.UserID) Operator {
	return Operator{
		user: &user,
	}
}

func FromIntegration(integration IntegrationID) Operator {
	return Operator{
		integration: &integration,
	}
}

func FromMachine() Operator {
	return Operator{
		isMachine: true,
	}
}

func (o Operator) User() *accountdomain.UserID {
	return o.user.CloneRef()
}

func (o Operator) Integration() *IntegrationID {
	return o.integration.CloneRef()
}

func (o Operator) Machine() bool {
	return o.isMachine
}

func (o Operator) Validate() bool {
	return !o.user.IsNil() || !o.integration.IsNil() || o.Machine()
}
