package test_suites

import (
	"context"
	"duolingo/libraries/connection_manager"
	"duolingo/libraries/connection_manager/test/fake"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ConnectionManagerTestSuite struct {
	suite.Suite
	ctx     context.Context
	cancel  context.CancelFunc
	builder *connection_manager.ConnectionBuilder
	manager *connection_manager.ConnectionManager
	proxy   *fake.FakeConnectionProxy
}

func TestConnectionManagerSuite(t *testing.T) {
	suite.Run(t, new(ConnectionManagerTestSuite))
}

func (s *ConnectionManagerTestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.proxy = fake.NewFakeConnectionProxy()
	s.builder = connection_manager.NewConnectionBuilder(context.Background()).
		SetConnectionDriver(s.proxy).
		SetURI("fake-uri").
		SetOperationRetryWait(5 * time.Millisecond).
		SetConnectionTimeOut(50 * time.Millisecond).
		SetOperationReadTimeOut(100 * time.Millisecond).
		SetOperationWriteTimeOut(100 * time.Millisecond)
	manager, err := s.builder.BuildConnectionManager()
	if err != nil {
		panic(err)
	}
	s.manager = manager
}

func (s *ConnectionManagerTestSuite) TearDownTest() {
	defer s.cancel()
	s.builder.Destroy()
	s.builder = nil
	s.manager = nil
}

func (s *ConnectionManagerTestSuite) TestAddClient() {
	client, err := s.builder.BuildClientAndRegisterToManager()
	if !s.Assert().NoError(err) || !s.Assert().NotNil(client) {
		return
	}
}

func (s *ConnectionManagerTestSuite) TestRemoveClient() {
	client, _ := s.builder.BuildClientAndRegisterToManager()
	s.manager.RemoveClient(client)
	s.Assert().Nil(client.GetConnection())
	s.Assert().Nil(s.manager.GetClientConnection(client))
}

func (s *ConnectionManagerTestSuite) TestGetClientConnection() {
	client, _ := s.builder.BuildClientAndRegisterToManager()
	s.Assert().NotNil(client.GetConnection())
	s.Assert().NotNil(s.manager.GetClientConnection(client))
	s.Assert().Equal(client.GetConnection(), s.manager.GetClientConnection(client))
}

func (s *ConnectionManagerTestSuite) TestConnectionReset() {
	var wg sync.WaitGroup
	var clientCount = 10
	var clients = make([]*connection_manager.Client, clientCount)
	var clientWork = func(ctx context.Context, conn any) error {
		if !s.proxy.IsNetworkUp() {
			return fake.ErrFakeNetworkFailure
		}
		return nil
	}
	// setup clients
	for i := range clientCount {
		client, _ := s.builder.BuildClientAndRegisterToManager()
		clients[i] = client
	}
	// verify client timeout on network failure
	s.proxy.SimulateNetworkFailure()
	wg.Add(clientCount)
	for i := range clientCount {
		go func() {
			defer wg.Done()
			timeout := clients[i].GetDefaultTimeOut() 
			err := clients[i].ExecuteClosure(timeout, clientWork)
			s.Assert().Equal(connection_manager.ErrClientOperationTimeout, err)
		}()
	}
	wg.Wait()
	// verify client works successful after connection recovered 
	s.proxy.SimulateNetworkRecovery()
	wg.Add(clientCount)
	for i := range clientCount {
		go func() {
			defer wg.Done()
			timeout := clients[i].GetDefaultTimeOut() 
			err := clients[i].ExecuteClosure(timeout, clientWork)
			s.Assert().NoError(err)
		}()
	}
	wg.Wait()
}
