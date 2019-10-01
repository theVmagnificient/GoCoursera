package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"strings"
	"time"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type ACLdata map[string][]string

type BizSrv struct {
}

type AdmSrv struct {
	stopCtx 	context.Context

	logChan		chan *Event 		// tracking single events logging
	logChanSub	chan chan *Event	// tracking that new logging subscriber is added
	logSubs		[]chan *Event		// for sending msg to each sub

	statChan 	chan *Event		/* Section for statistics */
	statChanSub	chan chan *Event
	statSubs	[]chan *Event
}


func (s *AdmSrv) Logging(_ *Nothing, stream Admin_LoggingServer) error {
	ch := make(chan *Event, 0)

	s.logChanSub <- ch

	for {
		select {
		case e := <- ch:
			err := stream.Send(e)
			if err != nil {
				return err
			}
		case <-s.stopCtx.Done():
			return nil
		}
	}
}

func (s *AdmSrv) Statistics(interval *StatInterval, stream Admin_StatisticsServer) error {
	ticker := time.NewTicker(time.Duration(interval.IntervalSeconds) * time.Second)

	stat := &Stat{
		ByMethod:        make(map[string]uint64),
		ByConsumer:      make(map[string]uint64),
	}
	ch := make(chan *Event, 0)

	s.statChanSub <- ch

	for {
		select {
		case st := <- ch:
			stat.ByMethod[st.Method] += 1
			stat.ByConsumer[st.Consumer] += 1
		case <-ticker.C:
			err := stream.Send(stat)
			if err != nil {
				return err
			}
			stat.ByMethod = make(map[string]uint64)
			stat.ByConsumer = make(map[string]uint64)
		case <- s.stopCtx.Done():
			return nil
		}
	}
}

type USrv struct {
	BizSrv
	AdmSrv
	ACLdata
}
// ---

func (s BizSrv) Add(ctx context.Context, _ *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (s BizSrv) Check(ctx context.Context, _ *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (s BizSrv) Test(ctx context.Context, _ *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}


func (s *USrv) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if err := s.checkAuth(ctx, info.FullMethod); err != nil {
		return nil, err
	}
	// Calls the handler
	h, err := handler(ctx, req)

	return h, err
}

func (s *USrv) StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := s.checkAuth(stream.Context(), info.FullMethod); err != nil {
		return err
	}

	// Calls the handler
	err := handler(srv, stream)

	return err
}


func (s *USrv) checkAuth(ctx context.Context, methodName string) error {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return status.Errorf(codes.Unauthenticated, "can't get metadata")
	}

	c, ok := md["consumer"]

	if !ok {
		return status.Errorf(codes.Unauthenticated, "can't get consumer field")
	}

	_, ok = s.ACLdata[c[0]]

	if !ok {
		return status.Errorf(codes.Unauthenticated, "No such method found")
	}

	partsMethod := strings.Split(methodName, "/")
	found := false

	for _, val := range s.ACLdata[c[0]] {
		partsCur := strings.Split(val, "/")

		if (partsCur[1] == partsMethod[1] && partsCur[2] == partsMethod[2]) ||
			(partsCur[1] == partsMethod[1]) && partsCur[2] == "*" {
			found = true
		}
	}
	if !found {
		return status.Errorf(codes.Unauthenticated, "can't get method by name")
	}

	s.logChan <- &Event{
		Consumer:    c[0],
		Method:      methodName,
		Host:        "127.0.0.1:100500",
	}

	s.statChan <- &Event{
		Consumer: c[0],
		Method:   methodName,
	}

	return nil
}

func StartMyMicroservice(ctx context.Context, listenAddr string, aclData string) error {
	var acl ACLdata

	err := json.Unmarshal([]byte(aclData), &acl)

	if err != nil {
		return fmt.Errorf("Cannot unpack acl data")
	}

	lis, err := net.Listen("tcp", listenAddr)

	if err != nil {
		log.Fatalln("Can't listen the port", err)
	}

	logChan := make(chan *Event, 0)
	logSubChan := make(chan chan *Event, 0)
	logSubs := make([]chan *Event, 0)

	statChan := make(chan *Event, 0)
	statSubChan := make(chan chan *Event, 0)
	statSubs := make([]chan *Event, 0)

	srv := USrv{
		BizSrv:  BizSrv{},
		AdmSrv:  AdmSrv{
			stopCtx:     ctx,
			logChan:     logChan,
			logChanSub:  logSubChan,
			logSubs:     logSubs,
			statChan:    statChan,
			statChanSub: statSubChan,
			statSubs:    statSubs,
		},
		ACLdata: acl,
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(srv.UnaryServerInterceptor),
		grpc.StreamInterceptor(srv.StreamServerInterceptor),
		)

	RegisterBizServer(server, srv)
	RegisterAdminServer(server, &srv)

	go func () {
		for {
			select {
			case newCh := <- srv.logChanSub: // new sub added
				srv.logSubs = append(srv.logSubs, newCh)
			case newEvent := <- srv.logChan: // send new event to all subs
				for _, ch := range srv.logSubs {
					ch <- newEvent
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	go func () {
		for {
			select {
			case newCh := <- srv.statChanSub: // new sub added
				srv.statSubs = append(srv.statSubs, newCh)
			case newEvent := <- srv.statChan: // send new event to all subs
				for _, ch := range srv.statSubs {
					ch <- newEvent
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// grpc server start logic (non-blocking)
	go func() {
		err = server.Serve(lis)
		if err != nil {
			log.Fatal("grpc server start failed", err)
		}
	}()

	// grpc server stop logic
	go func() {
		<-ctx.Done() // await until everything is done and the stop signal is acquired
		server.Stop()
	}()

	return nil
}