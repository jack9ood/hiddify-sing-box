package inbound

import (
	"context"
	"net"
	"os"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/mux"
	"github.com/sagernet/sing-box/common/tls"
	"github.com/sagernet/sing-box/common/uot"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/uap"
	"github.com/sagernet/sing-box/transport/v2ray"
	vmess "github.com/sagernet/sing-vmess"
	"github.com/sagernet/sing-vmess/packetaddr"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/auth"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

var (
	_ adapter.Inbound           = (*UAP)(nil)
	_ adapter.InjectableInbound = (*UAP)(nil)
)

type UAP struct {
	myInboundAdapter
	ctx       context.Context
	users     []option.UAPUser
	service   *uap.Service[int]
	tlsConfig tls.ServerConfig
	transport adapter.V2RayServerTransport
}

func NewUAP(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.UAPInboundOptions) (*UAP, error) {
	inbound := &UAP{
		myInboundAdapter: myInboundAdapter{
			protocol:      C.TypeUAP,
			network:       []string{N.NetworkTCP},
			ctx:           ctx,
			router:        uot.NewRouter(router, logger),
			logger:        logger,
			tag:           tag,
			listenOptions: options.ListenOptions,
		},
		ctx:   ctx,
		users: options.Users,
	}
	var err error
	inbound.router, err = mux.NewRouterWithOptions(inbound.router, logger, common.PtrValueOrDefault(options.Multiplex))
	if err != nil {
		return nil, err
	}
	service := uap.NewService[int](logger, adapter.NewUpstreamContextHandler(inbound.newConnection, inbound.newPacketConnection, inbound))
	service.UpdateUsers(common.MapIndexed(inbound.users, func(index int, _ option.UAPUser) int {
		return index
	}), common.Map(inbound.users, func(it option.UAPUser) string {
		return it.UUID
	}), common.Map(inbound.users, func(it option.UAPUser) string {
		return it.Flow
	}))
	inbound.service = service
	if options.TLS != nil {
		inbound.tlsConfig, err = tls.NewServer(ctx, logger, common.PtrValueOrDefault(options.TLS))
		if err != nil {
			return nil, err
		}
	}
	if options.Transport != nil {
		inbound.transport, err = v2ray.NewServerTransport(ctx, common.PtrValueOrDefault(options.Transport), inbound.tlsConfig, (*uapTransportHandler)(inbound))
		if err != nil {
			return nil, E.Cause(err, "create server transport: ", options.Transport.Type)
		}
	}
	inbound.connHandler = inbound
	return inbound, nil
}

func (h *UAP) Start() error {
	err := common.Start(
		h.service,
		h.tlsConfig,
	)
	if err != nil {
		return err
	}
	if h.transport == nil {
		return h.myInboundAdapter.Start()
	}
	if common.Contains(h.transport.Network(), N.NetworkTCP) {
		tcpListener, err := h.myInboundAdapter.ListenTCP()
		if err != nil {
			return err
		}
		go func() {
			sErr := h.transport.Serve(tcpListener)
			if sErr != nil && !E.IsClosed(sErr) {
				h.logger.Error("transport serve error: ", sErr)
			}
		}()
	}
	if common.Contains(h.transport.Network(), N.NetworkUDP) {
		udpConn, err := h.myInboundAdapter.ListenUDP()
		if err != nil {
			return err
		}
		go func() {
			sErr := h.transport.ServePacket(udpConn)
			if sErr != nil && !E.IsClosed(sErr) {
				h.logger.Error("transport serve error: ", sErr)
			}
		}()
	}
	return nil
}

func (h *UAP) Close() error {
	return common.Close(
		h.service,
		&h.myInboundAdapter,
		h.tlsConfig,
		h.transport,
	)
}

func (h *UAP) newTransportConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	h.injectTCP(conn, metadata)
	return nil
}

func (h *UAP) NewConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	var err error
	if h.tlsConfig != nil && h.transport == nil {
		conn, err = tls.ServerHandshake(ctx, conn, h.tlsConfig)
		if err != nil {
			return err
		}
	}
	return h.service.NewConnection(adapter.WithContext(log.ContextWithNewID(ctx), &metadata), conn, adapter.UpstreamMetadata(metadata))
}

func (h *UAP) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	return os.ErrInvalid
}

func (h *UAP) newConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	userIndex, loaded := auth.UserFromContext[int](ctx)
	if !loaded {
		return os.ErrInvalid
	}
	user := h.users[userIndex].Name
	if user == "" {
		user = F.ToString(userIndex)
	} else {
		metadata.User = user
	}
	h.logger.InfoContext(ctx, "[", user, "] inbound connection to ", metadata.Destination)
	return h.router.RouteConnection(ctx, conn, metadata)
}

func (h *UAP) newPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	userIndex, loaded := auth.UserFromContext[int](ctx)
	if !loaded {
		return os.ErrInvalid
	}
	user := h.users[userIndex].Name
	if user == "" {
		user = F.ToString(userIndex)
	} else {
		metadata.User = user
	}
	if metadata.Destination.Fqdn == packetaddr.SeqPacketMagicAddress {
		metadata.Destination = M.Socksaddr{}
		conn = packetaddr.NewConn(conn.(vmess.PacketConn), metadata.Destination)
		h.logger.InfoContext(ctx, "[", user, "] inbound packet addr connection")
	} else {
		h.logger.InfoContext(ctx, "[", user, "] inbound packet connection to ", metadata.Destination)
	}
	return h.router.RoutePacketConnection(ctx, conn, metadata)
}

var _ adapter.V2RayServerTransportHandler = (*uapTransportHandler)(nil)

type uapTransportHandler UAP

func (t *uapTransportHandler) NewConnection(ctx context.Context, conn net.Conn, metadata M.Metadata) error {
	return (*UAP)(t).newTransportConnection(ctx, conn, adapter.InboundContext{
		Source:      metadata.Source,
		Destination: metadata.Destination,
	})
}
