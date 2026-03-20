package chainbridge

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bsv8/BFTP/pkg/feepool/dual2of2"
	chainapi "github.com/bsv8/BSVChainAPI"
)

const DefaultPortBaseURL = "http://127.0.0.1:18222"

// FeePoolChainBaseURLEnv 用于跨进程共享同一个 Chain API 端口。
// 有值时优先走 PortClient；无值时退回当前进程内嵌 Manager。
const FeePoolChainBaseURLEnv = "BSV_CHAIN_API_URL"

type RouteConfig struct {
	Provider string
	Network  string
	Profile  string
}

func (c RouteConfig) Normalize() chainapi.Route {
	route := chainapi.Route{
		Provider: strings.TrimSpace(c.Provider),
		Network:  strings.TrimSpace(c.Network),
		Profile:  strings.TrimSpace(c.Profile),
	}.Normalize()
	if route.Provider == "" {
		route.Provider = chainapi.WhatsOnChainProvider
	}
	return route.Normalize()
}

type FeePoolChain struct {
	api     chainapi.API
	route   chainapi.Route
	baseURL string
}

func NewEmbeddedFeePoolChain(routeCfg RouteConfig, protectInterval time.Duration) (*FeePoolChain, error) {
	route := routeCfg.Normalize()
	manager, err := chainapi.NewManager(chainapi.Config{
		Routes: []chainapi.RouteConfig{
			{
				Provider: route.Provider,
				Network:  route.Network,
				Profile:  route.Profile,
				Protect: chainapi.ProtectConfig{
					MinInterval: protectInterval,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return &FeePoolChain{
		api:     manager,
		route:   route,
		baseURL: "embedded",
	}, nil
}

// ResolvePortBaseURLFromEnv 返回环境变量声明的共享 Chain API 地址。
// 这里只做归一化，不做连通性探测，避免启动阶段引入隐式阻塞。
func ResolvePortBaseURLFromEnv() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv(FeePoolChainBaseURLEnv)), "/")
}

// NewDefaultFeePoolChain 根据环境选择共享端口模式或进程内嵌模式。
func NewDefaultFeePoolChain(routeCfg RouteConfig, protectInterval time.Duration) (*FeePoolChain, error) {
	if u := ResolvePortBaseURLFromEnv(); u != "" {
		return NewPortFeePoolChain(u, routeCfg), nil
	}
	return NewEmbeddedFeePoolChain(routeCfg, protectInterval)
}

func NewPortFeePoolChain(baseURL string, routeCfg RouteConfig) *FeePoolChain {
	route := routeCfg.Normalize()
	u := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if u == "" {
		u = DefaultPortBaseURL
	}
	return &FeePoolChain{
		api:     chainapi.NewPortClient(u),
		route:   route,
		baseURL: u,
	}
}

func NewSharedHandler(routeCfg RouteConfig, protectInterval time.Duration) (http.Handler, error) {
	route := routeCfg.Normalize()
	manager, err := chainapi.NewManager(chainapi.Config{
		Routes: []chainapi.RouteConfig{
			{
				Provider: route.Provider,
				Network:  route.Network,
				Profile:  route.Profile,
				Protect: chainapi.ProtectConfig{
					MinInterval: protectInterval,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return chainapi.NewPortServer(manager).Handler(), nil
}

func (c *FeePoolChain) GetUTXOs(address string) ([]dual2of2.UTXO, error) {
	if c == nil || c.api == nil {
		return nil, fmt.Errorf("fee pool chain is nil")
	}
	items, err := c.api.GetUTXOsContext(context.Background(), c.route, strings.TrimSpace(address))
	if err != nil {
		return nil, err
	}
	out := make([]dual2of2.UTXO, 0, len(items))
	for _, item := range items {
		out = append(out, dual2of2.UTXO{
			TxID:  item.TxID,
			Vout:  item.Vout,
			Value: item.Value,
		})
	}
	return out, nil
}

func (c *FeePoolChain) GetTipHeight() (uint32, error) {
	if c == nil || c.api == nil {
		return 0, fmt.Errorf("fee pool chain is nil")
	}
	return c.api.GetTipHeightContext(context.Background(), c.route)
}

func (c *FeePoolChain) Broadcast(txHex string) (string, error) {
	if c == nil || c.api == nil {
		return "", fmt.Errorf("fee pool chain is nil")
	}
	return c.api.BroadcastContext(context.Background(), c.route, strings.TrimSpace(txHex))
}

func (c *FeePoolChain) BaseURL() string {
	if c == nil {
		return ""
	}
	return c.baseURL
}

func (c *FeePoolChain) Route() chainapi.Route {
	if c == nil {
		return chainapi.Route{}
	}
	return c.route
}
