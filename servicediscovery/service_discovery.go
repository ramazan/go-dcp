package servicediscovery

import (
	"fmt"
	"sort"
	"time"

	"github.com/Trendyol/go-dcp/wrapper"

	"github.com/Trendyol/go-dcp/config"

	"github.com/Trendyol/go-dcp/membership"

	"github.com/Trendyol/go-dcp/helpers"

	"github.com/Trendyol/go-dcp/logger"
)

type ServiceDiscovery interface {
	Add(service *Service)
	Remove(name string)
	RemoveAll()
	AssignLeader(leaderService *Service)
	RemoveLeader()
	ReassignLeader() error
	StartHeartbeat()
	StopHeartbeat()
	StartMonitor()
	StopMonitor()
	GetAll() []string
	SetInfo(memberNumber int, totalMembers int)
	BeLeader()
	DontBeLeader()
}

type serviceDiscovery struct {
	bus             helpers.Bus
	leaderService   *Service
	services        *wrapper.ConcurrentSwissMap[string, *Service]
	heartbeatTicker *time.Ticker
	monitorTicker   *time.Ticker
	info            *membership.Model
	config          *config.Dcp
	amILeader       bool
}

func (s *serviceDiscovery) Add(service *Service) {
	s.services.Store(service.Name, service)
}

func (s *serviceDiscovery) Remove(name string) {
	if service, ok := s.services.Load(name); ok {
		_ = service.Client.Close()

		s.services.Delete(name)
	}
}

func (s *serviceDiscovery) RemoveAll() {
	s.services.Range(func(name string, _ *Service) bool {
		s.services.Delete(name)

		return true
	})
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

func (s *serviceDiscovery) StartHeartbeat() {
	s.heartbeatTicker = time.NewTicker(5 * time.Second)

	go func() {
		for range s.heartbeatTicker.C {
			if s.leaderService != nil {
				err := s.leaderService.Client.Ping()
				if err != nil {
					logger.Log.Info("leader is down, health check failed for leader")

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

			s.services.Range(func(name string, service *Service) bool {
				err := service.Client.Ping()
				if err != nil {
					s.Remove(name)
					logger.Log.Info("client %s disconnected", name)
				}

				return true
			})
		}
	}()
}

func (s *serviceDiscovery) StopHeartbeat() {
	s.heartbeatTicker.Stop()
}

func (s *serviceDiscovery) StartMonitor() {
	s.monitorTicker = time.NewTicker(5 * time.Second)

	go func() {
		logger.Log.Info("service discovery will start after %v", s.config.Dcp.Group.Membership.RebalanceDelay)
		time.Sleep(s.config.Dcp.Group.Membership.RebalanceDelay)

		for range s.monitorTicker.C {
			if !s.amILeader {
				continue
			}

			names := s.GetAll()
			totalMembers := len(names) + 1

			s.SetInfo(1, totalMembers)

			for index, name := range names {
				if service, ok := s.services.Load(name); ok {
					if err := service.Client.Rebalance(index+2, totalMembers); err != nil {
						logger.Log.Error("rebalance failed for %s", name)
					}
				}
			}
		}
	}()
}

func (s *serviceDiscovery) StopMonitor() {
	s.monitorTicker.Stop()
}

func (s *serviceDiscovery) GetAll() []string {
	var names []string

	s.services.Range(func(name string, _ *Service) bool {
		names = append(names, name)

		return true
	})

	sort.Strings(names)

	return names
}

func (s *serviceDiscovery) SetInfo(memberNumber int, totalMembers int) {
	newInfo := &membership.Model{
		MemberNumber: memberNumber,
		TotalMembers: totalMembers,
	}

	if newInfo.IsChanged(s.info) {
		s.info = newInfo
		logger.Log.Info("new info arrived for member: %v/%v", memberNumber, totalMembers)
		s.bus.Emit(helpers.MembershipChangedBusEventName, newInfo)
	}
}

func NewServiceDiscovery(config *config.Dcp, bus helpers.Bus) ServiceDiscovery {
	return &serviceDiscovery{
		services: wrapper.CreateConcurrentSwissMap[string, *Service](0),
		bus:      bus,
		config:   config,
	}
}
