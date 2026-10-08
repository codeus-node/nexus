package nexus

import "github.com/codeus-node/fail"

type Runner interface {
	Start(config Config) fail.CustomError
}

func New(instance Runner) Runner {
	return instance
}
