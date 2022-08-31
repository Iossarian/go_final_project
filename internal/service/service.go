package service

import (
	"abf/internal/bucket"
	"abf/internal/service/pb"
	"context"
	"errors"
	"net"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	bs BucketStorage
	ls ListStorage
	pb.UnimplementedAntiBruteForceServer
}

type BucketStorage interface {
	Check(login, password, ip string) error
	Clear(login, password, ip string)
}

type ListStorage interface {
	AddWhiteNet(net string) error
	AddBlackNet(net string) error
	RemoveWhiteNet(net string) error
	RemoveBlackNet(net string) error
	Check(ip net.IP) (IPStatus, error)
}

type IPStatus int

const (
	Allowed IPStatus = iota + 1
	Rejected
	Undefined
)

func NewService(bs BucketStorage, ls ListStorage) *Service {
	return &Service{
		bs: bs,
		ls: ls,
	}
}

func (s *Service) AddWhiteNet(_ context.Context, request *pb.IpRequest) (*pb.Status, error) {
	_, ipNet, err := net.ParseCIDR(request.GetIp())
	if err != nil {
		return &pb.Status{Ok: &wrappers.BoolValue{Value: false}}, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = s.ls.AddWhiteNet(ipNet.String()); err != nil {
		return &pb.Status{Ok: &wrappers.BoolValue{Value: false}}, status.Error(codes.Internal, err.Error())
	}

	return &pb.Status{Ok: &wrappers.BoolValue{Value: true}}, nil
}

func (s *Service) AddBlackNet(_ context.Context, request *pb.IpRequest) (*pb.Status, error) {
	_, ipNet, err := net.ParseCIDR(request.GetIp())
	if err != nil {
		return &pb.Status{Ok: &wrappers.BoolValue{Value: false}}, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = s.ls.AddBlackNet(ipNet.String()); err != nil {
		return &pb.Status{Ok: &wrappers.BoolValue{Value: false}}, status.Error(codes.Internal, err.Error())
	}

	return &pb.Status{Ok: &wrappers.BoolValue{Value: true}}, nil
}

func (s *Service) RemoveWhiteNet(_ context.Context, request *pb.IpRequest) (*pb.Status, error) {
	_, ipNet, err := net.ParseCIDR(request.GetIp())
	if err != nil {
		return &pb.Status{Ok: &wrappers.BoolValue{Value: false}}, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = s.ls.RemoveWhiteNet(ipNet.String()); err != nil && !errors.Is(err, memcache.ErrCacheMiss) {
		return &pb.Status{Ok: &wrappers.BoolValue{Value: false}}, status.Error(codes.Internal, err.Error())
	}

	return &pb.Status{Ok: &wrappers.BoolValue{Value: true}}, nil
}

func (s *Service) RemoveBlackNet(_ context.Context, request *pb.IpRequest) (*pb.Status, error) {
	_, ipNet, err := net.ParseCIDR(request.GetIp())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = s.ls.RemoveBlackNet(ipNet.String()); err != nil && !errors.Is(err, memcache.ErrCacheMiss) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Status{Ok: &wrappers.BoolValue{Value: true}}, nil
}

func (s *Service) ClearBucket(_ context.Context, request *pb.CheckRequest) (*empty.Empty, error) {
	s.bs.Clear(request.GetLogin(), request.GetPassword(), request.GetIp())

	return &empty.Empty{}, nil
}

func (s *Service) Try(_ context.Context, request *pb.CheckRequest) (*pb.Status, error) {
	ip := net.ParseIP(request.GetIp())
	if ip == nil {
		return nil, status.Error(codes.InvalidArgument, "IP field is invalid")
	}

	statusIP, err := s.ls.Check(ip)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	switch statusIP {
	case Allowed:
		return &pb.Status{Ok: &wrappers.BoolValue{Value: true}}, nil
	case Rejected:
		return &pb.Status{Ok: &wrappers.BoolValue{Value: false}}, nil
	case Undefined:
		break
	}

	err = s.bs.Check(request.GetLogin(), request.GetPassword(), request.GetIp())
	if errors.Is(bucket.ErrRejected, err) {
		return &pb.Status{Ok: &wrappers.BoolValue{Value: false}}, nil
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Status{Ok: &wrappers.BoolValue{Value: true}}, nil
}
