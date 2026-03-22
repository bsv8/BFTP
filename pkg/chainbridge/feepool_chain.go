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

// FeePoolChainPlan 把读链路由和提交策略分开描述。
// 读链只需要一个稳定 route，提交可以按顺序尝试多个 route。
type FeePoolChainPlan struct {
	Routes         []chainapi.RouteConfig
	ReadRoute      chainapi.Route
	TxSubmitPolicy chainapi.TxSubmitPolicy
}

func (c RouteConfig) Normalize() chainapi.Route {
	return chainapi.Route{
		Provider: strings.TrimSpace(c.Provider),
		Network:  strings.TrimSpace(c.Network),
		Profile:  strings.TrimSpace(c.Profile),
	}.Normalize()
}

type FeePoolChain struct {
	manager        chainapi.CapabilityAPI
	readRoute      chainapi.Route
	txSubmitPolicy chainapi.TxSubmitPolicy
	baseURL        string
}

func NewEmbeddedFeePoolChain(routeCfg RouteConfig, protectInterval time.Duration) (*FeePoolChain, error) {
	plan, err := singleRouteFeePoolChainPlan(routeCfg, protectInterval)
	if err != nil {
		return nil, err
	}
	return NewEmbeddedFeePoolChainFromPlan(plan)
}

func NewEmbeddedFeePoolChainFromPlan(plan FeePoolChainPlan) (*FeePoolChain, error) {
	normalized, err := normalizeFeePoolChainPlan(plan)
	if err != nil {
		return nil, err
	}
	manager, err := chainapi.NewManager(chainapi.Config{Routes: normalized.Routes})
	if err != nil {
		return nil, err
	}
	capManager, err := chainapi.NewCapabilityManager(context.Background(), manager, feePoolCapabilityPlan(normalized.ReadRoute, normalized.TxSubmitPolicy))
	if err != nil {
		return nil, err
	}
	return newFeePoolChain(capManager, normalized.ReadRoute, normalized.TxSubmitPolicy, "embedded"), nil
}

// ResolvePortBaseURLFromEnv 返回环境变量声明的共享 Chain API 地址。
// 这里只做归一化，不做连通性探测，避免启动阶段引入隐式阻塞。
func ResolvePortBaseURLFromEnv() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv(FeePoolChainBaseURLEnv)), "/")
}

// NewDefaultFeePoolChain 根据环境选择共享端口模式或进程内嵌模式。
func NewDefaultFeePoolChain(routeCfg RouteConfig, protectInterval time.Duration) (*FeePoolChain, error) {
	if u := ResolvePortBaseURLFromEnv(); u != "" {
		return NewPortFeePoolChain(u, routeCfg)
	}
	return NewEmbeddedFeePoolChain(routeCfg, protectInterval)
}

func NewPortFeePoolChain(baseURL string, routeCfg RouteConfig) (*FeePoolChain, error) {
	plan, err := singleRouteFeePoolChainPlan(routeCfg, 0)
	if err != nil {
		return nil, err
	}
	return NewPortFeePoolChainFromPlan(baseURL, plan)
}

func NewPortFeePoolChainFromPlan(baseURL string, plan FeePoolChainPlan) (*FeePoolChain, error) {
	normalized, err := normalizeFeePoolChainPlan(plan)
	if err != nil {
		return nil, err
	}
	u := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if u == "" {
		u = DefaultPortBaseURL
	}
	client := chainapi.NewCapabilityPortClient(u)
	return newFeePoolChain(client, normalized.ReadRoute, normalized.TxSubmitPolicy, u), nil
}

func NewSharedHandler(routeCfg RouteConfig, protectInterval time.Duration) (http.Handler, error) {
	plan, err := singleRouteFeePoolChainPlan(routeCfg, protectInterval)
	if err != nil {
		return nil, err
	}
	return NewSharedHandlerFromPlan(plan)
}

func NewSharedHandlerFromPlan(plan FeePoolChainPlan) (http.Handler, error) {
	normalized, err := normalizeFeePoolChainPlan(plan)
	if err != nil {
		return nil, err
	}
	manager, err := chainapi.NewManager(chainapi.Config{Routes: normalized.Routes})
	if err != nil {
		return nil, err
	}
	capManager, err := chainapi.NewCapabilityManager(context.Background(), manager, feePoolCapabilityPlan(normalized.ReadRoute, normalized.TxSubmitPolicy))
	if err != nil {
		return nil, err
	}
	return chainapi.NewCapabilityPortServer(capManager).Handler(), nil
}

func (c *FeePoolChain) GetUTXOs(address string) ([]dual2of2.UTXO, error) {
	if c == nil || c.manager == nil {
		return nil, fmt.Errorf("fee pool chain is nil")
	}
	items, err := c.manager.GetUTXOsContext(context.Background(), strings.TrimSpace(address))
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
	if c == nil || c.manager == nil {
		return 0, fmt.Errorf("fee pool chain is nil")
	}
	return c.manager.GetTipHeightContext(context.Background())
}

func (c *FeePoolChain) Broadcast(txHex string) (string, error) {
	if c == nil || c.manager == nil {
		return "", fmt.Errorf("fee pool chain is nil")
	}
	// 广播由独立 router 负责主用/接盘，不再直接绑定单一路由。
	result, err := c.manager.SubmitTxContext(context.Background(), strings.TrimSpace(txHex))
	if err != nil {
		return "", err
	}
	return result.TxID, nil
}

func (c *FeePoolChain) BaseURL() string {
	if c == nil {
		return ""
	}
	return c.baseURL
}

// ReadRoute 返回费用池读链能力绑定的单一路由。
// 交易提交不走这里，而是由独立的 TxSubmitPolicy 决定。
func (c *FeePoolChain) ReadRoute() chainapi.Route {
	if c == nil {
		return chainapi.Route{}
	}
	return c.readRoute
}

// TxSubmitPolicy 返回广播所使用的固定顺序策略。
func (c *FeePoolChain) TxSubmitPolicy() chainapi.TxSubmitPolicy {
	if c == nil {
		return chainapi.TxSubmitPolicy{}
	}
	routes := append([]chainapi.Route(nil), c.txSubmitPolicy.Routes...)
	return chainapi.TxSubmitPolicy{Routes: routes}
}

func newFeePoolChain(manager chainapi.CapabilityAPI, readRoute chainapi.Route, txSubmitPolicy chainapi.TxSubmitPolicy, baseURL string) *FeePoolChain {
	normalizedPolicy := chainapi.TxSubmitPolicy{
		Routes: append([]chainapi.Route(nil), txSubmitPolicy.Routes...),
	}
	return &FeePoolChain{
		manager:        manager,
		readRoute:      readRoute.Normalize(),
		txSubmitPolicy: normalizedPolicy,
		baseURL:        baseURL,
	}
}

func singleRouteFeePoolChainPlan(routeCfg RouteConfig, protectInterval time.Duration) (FeePoolChainPlan, error) {
	route := routeCfg.Normalize()
	return FeePoolChainPlan{
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
		ReadRoute: route,
		TxSubmitPolicy: chainapi.TxSubmitPolicy{
			Routes: []chainapi.Route{route},
		},
	}, nil
}

func normalizeFeePoolChainPlan(plan FeePoolChainPlan) (FeePoolChainPlan, error) {
	if len(plan.Routes) == 0 {
		return FeePoolChainPlan{}, fmt.Errorf("fee pool chain routes are required")
	}
	routeConfigs := make([]chainapi.RouteConfig, 0, len(plan.Routes))
	seenRoutes := map[string]struct{}{}
	for _, rc := range plan.Routes {
		route := rc.Route()
		if route.Provider == "" {
			return FeePoolChainPlan{}, fmt.Errorf("route provider is required")
		}
		if route.Network == "" {
			return FeePoolChainPlan{}, fmt.Errorf("route network is required")
		}
		key := route.Key()
		if _, exists := seenRoutes[key]; exists {
			return FeePoolChainPlan{}, fmt.Errorf("duplicate route: %s", key)
		}
		seenRoutes[key] = struct{}{}
		routeConfigs = append(routeConfigs, rc)
	}
	readRoute := plan.ReadRoute.Normalize()
	if readRoute.Provider == "" {
		return FeePoolChainPlan{}, fmt.Errorf("read route provider is required")
	}
	if readRoute.Network == "" {
		return FeePoolChainPlan{}, fmt.Errorf("read route network is required")
	}
	if _, ok := seenRoutes[readRoute.Key()]; !ok {
		return FeePoolChainPlan{}, fmt.Errorf("read route is not present in route list: %s", readRoute.Key())
	}
	submitRoutes, err := normalizeChainBridgeSubmitRoutes(plan.TxSubmitPolicy.Routes)
	if err != nil {
		return FeePoolChainPlan{}, err
	}
	for _, route := range submitRoutes {
		if _, ok := seenRoutes[route.Key()]; !ok {
			return FeePoolChainPlan{}, fmt.Errorf("tx submit route is not present in route list: %s", route.Key())
		}
	}
	return FeePoolChainPlan{
		Routes:    routeConfigs,
		ReadRoute: readRoute,
		TxSubmitPolicy: chainapi.TxSubmitPolicy{
			Routes: submitRoutes,
		},
	}, nil
}

func feePoolCapabilityPlan(readRoute chainapi.Route, txSubmitPolicy chainapi.TxSubmitPolicy) chainapi.CapabilityPlan {
	return chainapi.SingleReadRouteCapabilityPlan(readRoute, txSubmitPolicy)
}

func normalizeChainBridgeSubmitRoutes(in []chainapi.Route) ([]chainapi.Route, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("tx submit policy routes are required")
	}
	out := make([]chainapi.Route, 0, len(in))
	seen := map[string]struct{}{}
	for _, route := range in {
		n := route.Normalize()
		if n.Provider == "" {
			return nil, fmt.Errorf("tx submit route provider is required")
		}
		if n.Network == "" {
			return nil, fmt.Errorf("tx submit route network is required")
		}
		key := n.Key()
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("duplicate tx submit route: %s", key)
		}
		seen[key] = struct{}{}
		out = append(out, n)
	}
	return out, nil
}
