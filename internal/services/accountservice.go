package services

import (
	"log"

	"github.com/galaxy-future/costpilot/internal/config"
	"github.com/galaxy-future/costpilot/internal/types"
)

type AccountService struct {
	cloudAccount []types.CloudAccount
}

func NewAccountService() *AccountService {
	s := &AccountService{}
	s.InitCloudAccounts()
	return s
}

func (s *AccountService) GetAccounts() []types.CloudAccount {
	return s.cloudAccount
}

// InitCloudAccounts
func (s *AccountService) InitCloudAccounts() {
	s.cloudAccount = config.GetGlobalConfig().CloudAccounts
	var a []string
	for _, v := range s.cloudAccount {
		a = append(a, v.Name)
	}
	log.Printf("I! get cloud account ready: %v\n", a)
	return
}
