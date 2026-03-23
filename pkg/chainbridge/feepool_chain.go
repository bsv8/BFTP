package chainbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bsv8/BFTP/pkg/feepool/dual2of2"
	chainapi "github.com/bsv8/BSVChainAPI"
	"github.com/bsv8/BSVChainAPI/whatsonchain"
)

const DefaultPortBaseURL = "http://127.0.0.1:18222"

// FeePoolChainBaseURLEnv 历史上用于共享 chainapi 端口。
// 现在统一退化为“记录一个逻辑来源字符串”，不再真正走 HTTP 中间层。
const FeePoolChainBaseURLEnv = "BSV_CHAIN_API_URL"

type RouteConfig struct {
	Provider string
	Network  string
	Profile  string
}

// FeePoolChainPlan 只描述 route 清单、一个读链 route，以及广播回退顺序。
// 设计约束：
// - 不再把提交策略放进 BSVChainAPI；
// - 费用池自己决定“按哪几条 route 顺序广播”。
type FeePoolChainPlan struct {
	Routes       []chainapi.RouteConfig
	ReadRoute    chainapi.Route
	SubmitRoutes []chainapi.Route
}

func (c RouteConfig) Normalize() chainapi.Route {
	return chainapi.Route{
		Provider: strings.TrimSpace(c.Provider),
		Network:  strings.TrimSpace(c.Network),
		Profile:  strings.TrimSpace(c.Profile),
	}.Normalize()
}

type FeePoolChain struct {
	manager      *chainapi.Manager
	readRoute    chainapi.Route
	submitRoutes []chainapi.Route
	baseURL      string
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
	return newFeePoolChain(manager, normalized.ReadRoute, normalized.SubmitRoutes, "embedded"), nil
}

func ResolvePortBaseURLFromEnv() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv(FeePoolChainBaseURLEnv)), "/")
}

// NewDefaultFeePoolChain 现在始终返回进程内直连的 route 客户端；
// 若环境变量有值，仅把它保留到 BaseURL 里，方便日志/状态页继续显示来源。
func NewDefaultFeePoolChain(routeCfg RouteConfig, protectInterval time.Duration) (*FeePoolChain, error) {
	chain, err := NewEmbeddedFeePoolChain(routeCfg, protectInterval)
	if err != nil {
		return nil, err
	}
	if u := ResolvePortBaseURLFromEnv(); u != "" {
		chain.baseURL = u
	}
	return chain, nil
}

// NewPortFeePoolChain 退化为“带一个来源标签的嵌入式直连链客户端”。
// 这样旧代码不用改调用点，但不再经过共享 HTTP 中间层。
func NewPortFeePoolChain(baseURL string, routeCfg RouteConfig) (*FeePoolChain, error) {
	plan, err := singleRouteFeePoolChainPlan(routeCfg, 0)
	if err != nil {
		return nil, err
	}
	return NewPortFeePoolChainFromPlan(baseURL, plan)
}

func NewPortFeePoolChainFromPlan(baseURL string, plan FeePoolChainPlan) (*FeePoolChain, error) {
	chain, err := NewEmbeddedFeePoolChainFromPlan(plan)
	if err != nil {
		return nil, err
	}
	u := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if u != "" {
		chain.baseURL = u
	}
	return chain, nil
}

// NewSharedHandler 仅保留历史入口，避免 e2e/工具层一次性炸掉。
// 当前不再承载真实 chainapi 代理，只返回一个可探活的占位 handler。
func NewSharedHandler(routeCfg RouteConfig, protectInterval time.Duration) (http.Handler, error) {
	_, err := singleRouteFeePoolChainPlan(routeCfg, protectInterval)
	if err != nil {
		return nil, err
	}
	return NewSharedHandlerFromPlan(FeePoolChainPlan{})
}

func NewSharedHandlerFromPlan(_ FeePoolChainPlan) (http.Handler, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"service": "feepool-chain-compat",
			"mode":    "direct-route",
		})
	})
	return mux, nil
}

func (c *FeePoolChain) GetUTXOs(address string) ([]dual2of2.UTXO, error) {
	if c == nil || c.manager == nil {
		return nil, fmt.Errorf("fee pool chain is nil")
	}
	address = strings.TrimSpace(address)
	client, err := c.manager.Open(c.readRoute)
	if err != nil {
		return nil, err
	}
	items, err := getUTXOsByRouteClient(context.Background(), client, address)
	if err != nil {
		return nil, err
	}
	out := make([]dual2of2.UTXO, 0, len(items))
	for _, item := range items {
		out = append(out, dual2of2.UTXO{TxID: item.TxID, Vout: item.Vout, Value: item.Value})
	}
	return out, nil
}

func (c *FeePoolChain) GetTipHeight() (uint32, error) {
	if c == nil || c.manager == nil {
		return 0, fmt.Errorf("fee pool chain is nil")
	}
	client, err := c.manager.Open(c.readRoute)
	if err != nil {
		return 0, err
	}
	return getTipByRouteClient(context.Background(), client)
}

func (c *FeePoolChain) Broadcast(txHex string) (string, error) {
	if c == nil || c.manager == nil {
		return "", fmt.Errorf("fee pool chain is nil")
	}
	txHex = strings.TrimSpace(txHex)
	if txHex == "" {
		return "", fmt.Errorf("tx_hex is required")
	}
	var lastErr error
	for _, route := range c.submitRoutes {
		client, err := c.manager.Open(route)
		if err != nil {
			lastErr = err
			continue
		}
		txid, err := broadcastByRouteClient(context.Background(), client, txHex)
		if err == nil {
			return txid, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		return "", fmt.Errorf("submit routes are empty")
	}
	return "", lastErr
}

func (c *FeePoolChain) BaseURL() string {
	if c == nil {
		return ""
	}
	return c.baseURL
}

func (c *FeePoolChain) ReadRoute() chainapi.Route {
	if c == nil {
		return chainapi.Route{}
	}
	return c.readRoute
}

func (c *FeePoolChain) SubmitRoutes() []chainapi.Route {
	if c == nil {
		return nil
	}
	return append([]chainapi.Route(nil), c.submitRoutes...)
}

func newFeePoolChain(manager *chainapi.Manager, readRoute chainapi.Route, submitRoutes []chainapi.Route, baseURL string) *FeePoolChain {
	return &FeePoolChain{
		manager:      manager,
		readRoute:    readRoute.Normalize(),
		submitRoutes: append([]chainapi.Route(nil), submitRoutes...),
		baseURL:      baseURL,
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
				Protect:  chainapi.ProtectConfig{MinInterval: protectInterval},
			},
		},
		ReadRoute:    route,
		SubmitRoutes: []chainapi.Route{route},
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
	if !routeSupportsRead(readRoute.Provider) {
		return FeePoolChainPlan{}, fmt.Errorf("read route does not support fee pool reads: %s", readRoute.Key())
	}
	submitRoutes, err := normalizeChainBridgeSubmitRoutes(plan.SubmitRoutes)
	if err != nil {
		return FeePoolChainPlan{}, err
	}
	for _, route := range submitRoutes {
		if _, ok := seenRoutes[route.Key()]; !ok {
			return FeePoolChainPlan{}, fmt.Errorf("submit route is not present in route list: %s", route.Key())
		}
	}
	return FeePoolChainPlan{
		Routes:       routeConfigs,
		ReadRoute:    readRoute,
		SubmitRoutes: submitRoutes,
	}, nil
}

func normalizeChainBridgeSubmitRoutes(in []chainapi.Route) ([]chainapi.Route, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("submit routes are required")
	}
	out := make([]chainapi.Route, 0, len(in))
	seen := map[string]struct{}{}
	for _, route := range in {
		n := route.Normalize()
		if n.Provider == "" {
			return nil, fmt.Errorf("submit route provider is required")
		}
		if n.Network == "" {
			return nil, fmt.Errorf("submit route network is required")
		}
		key := n.Key()
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("duplicate submit route: %s", key)
		}
		seen[key] = struct{}{}
		out = append(out, n)
	}
	return out, nil
}

func routeSupportsRead(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case chainapi.WhatsOnChainProvider:
		return true
	default:
		return false
	}
}

func getUTXOsByRouteClient(ctx context.Context, client *chainapi.RouteClient, address string) ([]whatsonchain.UTXO, error) {
	switch strings.ToLower(strings.TrimSpace(client.Provider())) {
	case chainapi.WhatsOnChainProvider:
		wocClient, err := client.WhatsOnChain()
		if err != nil {
			return nil, err
		}
		return wocClient.GetAddressConfirmedUnspent(ctx, address)
	default:
		return nil, fmt.Errorf("route %s does not support GetUTXOs", client.Route().Key())
	}
}

func getTipByRouteClient(ctx context.Context, client *chainapi.RouteClient) (uint32, error) {
	switch strings.ToLower(strings.TrimSpace(client.Provider())) {
	case chainapi.WhatsOnChainProvider:
		wocClient, err := client.WhatsOnChain()
		if err != nil {
			return 0, err
		}
		return wocClient.GetChainInfo(ctx)
	default:
		return 0, fmt.Errorf("route %s does not support GetTipHeight", client.Route().Key())
	}
}

func broadcastByRouteClient(ctx context.Context, client *chainapi.RouteClient, txHex string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(client.Provider())) {
	case chainapi.WhatsOnChainProvider:
		wocClient, err := client.WhatsOnChain()
		if err != nil {
			return "", err
		}
		return wocClient.PostTxRaw(ctx, txHex)
	case chainapi.GorillaPoolARCProvider:
		submitClient, err := client.GorillaPoolARC()
		if err != nil {
			return "", err
		}
		return submitClient.PostTxRaw(ctx, txHex)
	case chainapi.TAALARCProvider:
		submitClient, err := client.TAALARC()
		if err != nil {
			return "", err
		}
		return submitClient.PostTxRaw(ctx, txHex)
	case chainapi.TAALLegacyProvider:
		submitClient, err := client.TAALLegacy()
		if err != nil {
			return "", err
		}
		return submitClient.PostTxRaw(ctx, txHex)
	default:
		return "", fmt.Errorf("route %s does not support Broadcast", client.Route().Key())
	}
}
