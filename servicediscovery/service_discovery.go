package servicediscovery

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Trendyol/go-dcp-client/helpers"

	"github.com/Trendyol/go-dcp-client/logger"

	"github.com/Trendyol/go-dcp-client/membership/info"
)

type ServiceDiscovery interface {
	Add(service *Service)
	Remove(name string)
	RemoveAll()
	AssignLeader(leaderService *Service)
	RemoveLeader()
	ReassignLeader() error
	StartHealthCheck()
	StopHealthCheck()
	StartRebalance()
	StopRebalance()
	GetAll() []string
	SetInfo(memberNumber int, totalMembers int)
	BeLeader()
	DontBeLeader()
}

type serviceDiscovery struct {
	infoHandler         info.Handler
	leaderService       *Service
	services            map[string]*Service
	healthCheckSchedule *time.Ticker
	rebalanceSchedule   *time.Ticker
	info                *info.Model
	servicesLock        *sync.RWMutex
	amILeader           bool
	config              *helpers.Config
}

func (s *serviceDiscovery) Add(service *Service) {
	s.services[service.Name] = service
}

func (s *serviceDiscovery) Remove(name string) {
	if _, ok := s.services[name]; ok {
		_ = s.services[name].Client.Close()

		delete(s.services, name)
	}
}

func (s *serviceDiscovery) RemoveAll() {
	for name := range s.services {
		s.Remove(name)
	}
}

func (s *serviceDiscovery) BeLeader() {
	s.amILeader = true
}

func (s *serviceDiscovery) DontBeLeader() {
	s.amILeader = false
}

func (s *serviceDiscovery) AssignLeader(leaderService *Service) {
	s.leaderService = leaderService
}

func (s *serviceDiscovery) RemoveLeader() {
	if s.leaderService == nil {
		return
	}

	_ = s.leaderService.Client.Close()

	s.leaderService = nil
}

func (s *serviceDiscovery) ReassignLeader() error {
	if s.leaderService == nil {
		return fmt.Errorf("leader is not assigned")
	}

	err := s.leaderService.Client.Reconnect()

	if err == nil {
		err = s.leaderService.Client.Register()
	}

	return err
}

func (s *serviceDiscovery) StartHealthCheck() {
	s.healthCheckSchedule = time.NewTicker(3 * time.Second)

	go func() {
		for range s.healthCheckSchedule.C {
			s.servicesLock.Lock()

			if s.leaderService != nil {
				err := s.leaderService.Client.Ping()
				if err != nil {
					logger.Error(fmt.Errorf("leader is down"), "health check failed for leader")

					tempLeaderService := s.leaderService

					if err := s.ReassignLeader(); err != nil {
						if tempLeaderService != s.leaderService {
							_ = tempLeaderService.Client.Close()
						} else {
							s.RemoveLeader()
						}
					}
				}
			}

			for name, service := range s.services {
				err := service.Client.Ping()
				if err != nil {
					s.Remove(name)

					logger.Debug("client %s disconnected", name)
				}
			}

			s.servicesLock.Unlock()
		}
	}()
}

func (s *serviceDiscovery) StopHealthCheck() {
	s.healthCheckSchedule.Stop()
}

func (s *serviceDiscovery) StartRebalance() {
	s.rebalanceSchedule = time.NewTicker(3 * time.Second)

	go func() {
		time.Sleep(s.config.Dcp.Group.Membership.RebalanceDelay)

		for range s.rebalanceSchedule.C {
			if !s.amILeader {
				continue
			}

			s.servicesLock.RLock()

			names := s.GetAll()
			totalMembers := len(names) + 1

			s.SetInfo(1, totalMembers)

			for index, name := range names {
				if service, ok := s.services[name]; ok {
					if err := service.Client.Rebalance(index+2, totalMembers); err != nil {
						logger.Error(err, "rebalance failed for %s", name)
					}
				}
			}

			s.servicesLock.RUnlock()
		}
	}()
}

func (s *serviceDiscovery) StopRebalance() {
	s.rebalanceSchedule.Stop()
}

func (s *serviceDiscovery) GetAll() []string {
	names := make([]string, 0, len(s.services))

	for name := range s.services {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func (s *serviceDiscovery) SetInfo(memberNumber int, totalMembers int) {
	newInfo := &info.Model{
		MemberNumber: memberNumber,
		TotalMembers: totalMembers,
	}

	if newInfo.IsChanged(s.info) {
		s.info = newInfo

		logger.Debug("new info arrived for member: %v/%v", memberNumber, totalMembers)

		s.infoHandler.OnModelChange(newInfo)
	}
}

func NewServiceDiscovery(config *helpers.Config, infoHandler info.Handler) ServiceDiscovery {
	return &serviceDiscovery{
		services:     make(map[string]*Service),
		infoHandler:  infoHandler,
		servicesLock: &sync.RWMutex{},
		config:       config,
	}
}
