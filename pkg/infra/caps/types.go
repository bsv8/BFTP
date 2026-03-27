package caps

import (
	"fmt"
	"strings"

	ncall "github.com/bsv8/BFTP/pkg/infra/ncall"
)

type PublicCapability struct {
	ID      string
	Version uint32
}

// ModuleSpec 描述一个模块在角色装配时占用的公开合同面。
// v1 故意只收口我们现在真正要检查的几类声明：
// - internal ability
// - public capability
// - node.call route
// - 独立 pproto proto
// 不重新引入旧 capability 大包里的“完整对象模型”。
type ModuleSpec struct {
	InternalAbility  string
	PublicCapability *PublicCapability
	Routes           []string
	Protos           []string
	HTTPPaths        []string
}

type Bundle struct {
	Modules            []ModuleSpec
	InternalAbilities  []string
	PublicCapabilities []PublicCapability
	Routes             []string
	Protos             []string
	HTTPPaths          []string
}

// Assemble 负责把角色模块声明装成一个可校验 bundle。
// 设计意图：
// - 角色仍然显式声明“装了哪些模块”；
// - 装配器只做汇总、冲突检查、生成能力快照；
// - 不做隐式注册，不回到旧的全局注册中心。
func Assemble(modules ...ModuleSpec) (Bundle, error) {
	bundle := Bundle{
		Modules:            make([]ModuleSpec, 0, len(modules)),
		InternalAbilities:  make([]string, 0, len(modules)),
		PublicCapabilities: make([]PublicCapability, 0, len(modules)),
		Routes:             make([]string, 0, len(modules)),
		Protos:             make([]string, 0, len(modules)),
		HTTPPaths:          make([]string, 0, len(modules)),
	}
	abilityOwners := make(map[string]string, len(modules))
	capabilityOwners := make(map[string]string, len(modules))
	routeOwners := make(map[string]string, len(modules))
	protoOwners := make(map[string]string, len(modules))
	httpPathOwners := make(map[string]string, len(modules))

	for idx, raw := range modules {
		module, owner, err := normalizeModuleSpec(idx, raw)
		if err != nil {
			return Bundle{}, err
		}
		if previous, ok := abilityOwners[module.InternalAbility]; ok {
			return Bundle{}, fmt.Errorf("internal ability conflict: %q declared by %s and %s", module.InternalAbility, previous, owner)
		}
		abilityOwners[module.InternalAbility] = owner
		bundle.Modules = append(bundle.Modules, module)
		bundle.InternalAbilities = append(bundle.InternalAbilities, module.InternalAbility)

		if module.PublicCapability != nil {
			capability := *module.PublicCapability
			key := capabilityKey(capability)
			if previous, ok := capabilityOwners[key]; ok {
				return Bundle{}, fmt.Errorf("public capability conflict: %q declared by %s and %s", key, previous, owner)
			}
			capabilityOwners[key] = owner
			bundle.PublicCapabilities = append(bundle.PublicCapabilities, capability)
		}

		for _, route := range module.Routes {
			if previous, ok := routeOwners[route]; ok {
				return Bundle{}, fmt.Errorf("route conflict: %q declared by %s and %s", route, previous, owner)
			}
			routeOwners[route] = owner
			bundle.Routes = append(bundle.Routes, route)
		}
		for _, proto := range module.Protos {
			if previous, ok := protoOwners[proto]; ok {
				return Bundle{}, fmt.Errorf("proto conflict: %q declared by %s and %s", proto, previous, owner)
			}
			protoOwners[proto] = owner
			bundle.Protos = append(bundle.Protos, proto)
		}
		for _, httpPath := range module.HTTPPaths {
			if previous, ok := httpPathOwners[httpPath]; ok {
				return Bundle{}, fmt.Errorf("http path conflict: %q declared by %s and %s", httpPath, previous, owner)
			}
			httpPathOwners[httpPath] = owner
			bundle.HTTPPaths = append(bundle.HTTPPaths, httpPath)
		}
	}
	return bundle, nil
}

// MustAssemble 用于角色侧静态 bundle 定义。
// bundle 结构是代码常量的一部分，若声明冲突，应该在启动时立刻失败。
func MustAssemble(modules ...ModuleSpec) Bundle {
	bundle, err := Assemble(modules...)
	if err != nil {
		panic(err)
	}
	return bundle
}

// ShowBody 根据 bundle 中自动汇总出的 public capability 生成标准返回体。
func (b Bundle) ShowBody(nodePubkeyHex string) ncall.CapabilitiesShowBody {
	return BuildShowBody(nodePubkeyHex, b.PublicCapabilities)
}

func BuildShowBody(nodePubkeyHex string, items []PublicCapability) ncall.CapabilitiesShowBody {
	body := ncall.CapabilitiesShowBody{
		NodePubkeyHex: strings.ToLower(strings.TrimSpace(nodePubkeyHex)),
		Capabilities:  make([]*ncall.CapabilityItem, 0, len(items)),
	}
	for _, item := range items {
		if strings.TrimSpace(item.ID) == "" {
			continue
		}
		body.Capabilities = append(body.Capabilities, &ncall.CapabilityItem{
			ID:      strings.TrimSpace(item.ID),
			Version: item.Version,
		})
	}
	return body
}

func normalizeModuleSpec(index int, module ModuleSpec) (ModuleSpec, string, error) {
	module.InternalAbility = strings.TrimSpace(module.InternalAbility)
	if module.InternalAbility == "" {
		return ModuleSpec{}, "", fmt.Errorf("module #%d internal ability required", index+1)
	}
	owner := module.InternalAbility
	routes, err := normalizeStringList(owner, "route", module.Routes)
	if err != nil {
		return ModuleSpec{}, "", err
	}
	module.Routes = routes
	protos, err := normalizeStringList(owner, "proto", module.Protos)
	if err != nil {
		return ModuleSpec{}, "", err
	}
	module.Protos = protos
	httpPaths, err := normalizeStringList(owner, "http path", module.HTTPPaths)
	if err != nil {
		return ModuleSpec{}, "", err
	}
	module.HTTPPaths = httpPaths
	if module.PublicCapability != nil {
		capability := PublicCapability{
			ID:      strings.TrimSpace(module.PublicCapability.ID),
			Version: module.PublicCapability.Version,
		}
		if capability.ID == "" {
			return ModuleSpec{}, "", fmt.Errorf("%s public capability id required", owner)
		}
		module.PublicCapability = &capability
	}
	return module, owner, nil
}

func normalizeStringList(owner string, kind string, items []string) ([]string, error) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			return nil, fmt.Errorf("%s %s required", owner, kind)
		}
		if _, exists := seen[value]; exists {
			return nil, fmt.Errorf("%s %s duplicate: %q", owner, kind, value)
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out, nil
}

func capabilityKey(capability PublicCapability) string {
	return fmt.Sprintf("%s@%d", capability.ID, capability.Version)
}
