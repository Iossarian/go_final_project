package integration

import (
	"abf/internal/service/pb"
	"context"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"strconv"
	"testing"
)

type ServiceSuite struct {
	suite.Suite

	ctx        context.Context
	conn       *grpc.ClientConn
	grpcClient pb.AntiBruteForceClient
}

func (s *ServiceSuite) SetupSuite() {
	var err error

	s.conn, err = grpc.Dial("abf:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	s.grpcClient = pb.NewAntiBruteForceClient(s.conn)
	s.ctx = context.Background()
}

func (s *ServiceSuite) TearDownSuite() {
	s.Require().NoError(s.conn.Close())
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) TestAddNet() {
	s.Run("add to white list", func() {
		req := &pb.IpRequest{Ip: "127.0.0.1/25"}
		res, err := s.grpcClient.AddWhiteNet(s.ctx, req)

		s.Require().NoError(err)
		s.Require().True(res.GetOk().GetValue())

		_, err = s.grpcClient.RemoveWhiteNet(s.ctx, req)

		s.Require().NoError(err)
	})

	s.Run("add to black list", func() {
		req := &pb.IpRequest{Ip: "127.0.0.1/26"}
		res, err := s.grpcClient.AddWhiteNet(s.ctx, req)

		s.Require().NoError(err)
		s.Require().True(res.GetOk().GetValue())

		_, err = s.grpcClient.RemoveBlackNet(s.ctx, req)

		s.Require().NoError(err)
	})
}

func (s *ServiceSuite) TestRemoveNet() {
	s.Run("remove from white list", func() {
		req := &pb.IpRequest{Ip: "127.0.0.1/25"}
		res, err := s.grpcClient.RemoveWhiteNet(s.ctx, req)

		s.Require().NoError(err)
		s.Require().True(res.GetOk().GetValue())
	})

	s.Run("remove from black list", func() {
		req := &pb.IpRequest{Ip: "127.0.0.1/26"}
		res, err := s.grpcClient.RemoveBlackNet(s.ctx, req)

		s.Require().NoError(err)
		s.Require().True(res.GetOk().GetValue())
	})
}

func (s *ServiceSuite) TestBucketClearing() {
	s.Run("test clearing", func() {
		login := "testLogin"
		password := "testPassword"
		IP := "197.10.0.1"
		req := &pb.CheckRequest{Login: login, Password: password, Ip: IP}

		_, err := s.grpcClient.ClearBucket(s.ctx, req)
		s.Require().NoError(err)

		for i := 1; i <= 10; i++ {
			res, err := s.grpcClient.Try(s.ctx, req)
			s.Require().NoError(err)
			if i <= 10 {
				s.Require().True(res.GetOk().GetValue())
			} else {
				s.Require().False(res.GetOk().GetValue())
			}
		}
		_, err = s.grpcClient.ClearBucket(s.ctx, req)
		s.Require().NoError(err)

		res, err := s.grpcClient.Try(s.ctx, req)
		s.Require().NoError(err)
		s.Require().True(res.GetOk().GetValue())
	})
}

func (s *ServiceSuite) TestChecking() {
	login := "testLogin"
	password := "testPassword"
	IP := "192.161.10.1"
	req := &pb.CheckRequest{Login: login, Password: password, Ip: IP}
	s.Run("test success", func() {
		_, err := s.grpcClient.ClearBucket(s.ctx, req)
		s.Require().NoError(err)

		res, err := s.grpcClient.Try(s.ctx, req)

		s.Require().NoError(err)
		s.Require().True(res.GetOk().Value)

		_, err = s.grpcClient.ClearBucket(s.ctx, req)
		s.Require().NoError(err)
	})

	s.Run("test fail", func() {
		_, err := s.grpcClient.ClearBucket(s.ctx, req)
		s.Require().NoError(err)

		for i := 1; i <= 11; i++ {
			res, err := s.grpcClient.Try(s.ctx, req)
			s.Require().NoError(err)
			if i > 11 {
				s.Require().False(res.GetOk().GetValue())
			}
		}

		_, err = s.grpcClient.ClearBucket(s.ctx, req)
		s.Require().NoError(err)
	})

	s.Run("test success with white ip", func() {
		_, err := s.grpcClient.ClearBucket(s.ctx, req)
		s.Require().NoError(err)

		req := &pb.IpRequest{Ip: IP + "/25"}
		res, err := s.grpcClient.AddWhiteNet(s.ctx, req)

		s.Require().NoError(err)
		s.Require().True(res.GetOk().GetValue())

		for i := 1; i <= 1100; i++ {
			login = "testLogin" + strconv.Itoa(i)
			password = "testPassword" + strconv.Itoa(i)

			req := &pb.CheckRequest{Login: login, Password: password, Ip: IP}
			res, err := s.grpcClient.Try(s.ctx, req)

			s.Require().NoError(err)
			s.Require().True(res.GetOk().GetValue())
		}

		req = &pb.IpRequest{Ip: IP + "/25"}
		res, err = s.grpcClient.RemoveWhiteNet(s.ctx, req)

		s.Require().NoError(err)
		s.Require().True(res.GetOk().GetValue())
	})

	s.Run("test fail with black ip", func() {
		_, err := s.grpcClient.ClearBucket(s.ctx, req)
		s.Require().NoError(err)

		req := &pb.IpRequest{Ip: IP + "/25"}
		res, err := s.grpcClient.AddBlackNet(s.ctx, req)

		s.Require().NoError(err)
		s.Require().True(res.GetOk().GetValue())

		for i := 1; i <= 2; i++ {
			login = "testLogin" + strconv.Itoa(i)
			password = "testPassword" + strconv.Itoa(i)

			req := &pb.CheckRequest{Login: login, Password: password, Ip: IP}
			res, err := s.grpcClient.Try(s.ctx, req)

			s.Require().NoError(err)
			s.Require().False(res.GetOk().GetValue())
		}

		req = &pb.IpRequest{Ip: IP + "/25"}
		res, err = s.grpcClient.RemoveBlackNet(s.ctx, req)

		s.Require().NoError(err)
		s.Require().True(res.GetOk().GetValue())
	})
}
