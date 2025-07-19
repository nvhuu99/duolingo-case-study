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

func NewConnectionManagerTestSuite() *ConnectionManagerTestSuite {
	return &ConnectionManagerTestSuite{}
}

func TestConnectionManagerSuite(t *testing.T) {
	suite.Run(t, new(ConnectionManagerTestSuite))
}

func (s *ConnectionManagerTestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	args := connection_manager.DefaultConnectionArgs()
	args.SetConnectionTimeout(10 * time.Millisecond)
	args.SetRetryWait(5 * time.Millisecond)

	s.proxy = fake.NewFakeConnectionProxy()

	s.builder = connection_manager.NewConnectionBuilder(context.Background())
	s.builder.SetConnectionArgs(args)
	s.builder.SetConnectionProxy(s.proxy)

	s.manager = s.builder.GetConnectionManager()
}

func (s *ConnectionManagerTestSuite) TearDownTest() {
	s.builder.Destroy()
	s.cancel()
	s.builder = nil
	s.manager = nil
	s.proxy = nil
}

func (s *ConnectionManagerTestSuite) TestAddClient() {
	client := s.builder.BuildClientAndRegisterToManager()
	s.Assert().NotNil(client)
}

func (s *ConnectionManagerTestSuite) TestRemoveClient() {
	client := s.builder.BuildClientAndRegisterToManager()
	s.manager.RemoveClient(client)
	s.Assert().Nil(client.GetConnection())
	s.Assert().Nil(s.manager.GetClientConnection(client))
}

func (s *ConnectionManagerTestSuite) TestGetClientConnection() {
	client := s.builder.BuildClientAndRegisterToManager()
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
		clients[i] = s.builder.BuildClientAndRegisterToManager()
	}
	// verify client timeout on network failure
	s.proxy.SimulateNetworkFailure()
	wg.Add(clientCount)
	for i := range clientCount {
		go func() {
			defer wg.Done()
			err := clients[i].ExecuteClosure(20*time.Millisecond, clientWork)

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
			err := clients[i].ExecuteClosure(20*time.Millisecond, clientWork)

			s.Assert().NoError(err)
		}()
	}
	wg.Wait()
}
