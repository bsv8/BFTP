package dual2of2

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// Libp2pMarshalPubHexToSecpCompressedHex 兼容两种输入并统一输出 secp256k1 压缩公钥(hex，33字节)：
// - libp2p MarshalPublicKey(hex)
// - secp256k1 压缩公钥(hex)
// 这是把 p2prpc 的 sender_pubkey 语义，桥接到 KeymasterMultisigPool 的 BSV 公钥语义上。
func Libp2pMarshalPubHexToSecpCompressedHex(marshalPubHex string) (string, error) {
	marshalPubHex = strings.TrimSpace(marshalPubHex)
	if marshalPubHex == "" {
		return "", fmt.Errorf("pubkey hex required")
	}
	b, err := hex.DecodeString(marshalPubHex)
	if err != nil {
		return "", fmt.Errorf("decode pubkey hex: %w", err)
	}
	pub, err := crypto.UnmarshalPublicKey(b)
	if err == nil {
		raw, err := pub.Raw()
		if err != nil {
			return "", fmt.Errorf("extract raw pubkey: %w", err)
		}
		// KeymasterMultisigPool 需要 secp256k1 压缩公钥（33字节）。
		if len(raw) == 33 {
			return strings.ToLower(hex.EncodeToString(raw)), nil
		}
		return "", fmt.Errorf("unexpected raw pubkey length: %d (expect 33)", len(raw))
	}
	// 兼容已经是 secp256k1 压缩公钥的输入。
	pub, err = crypto.UnmarshalSecp256k1PublicKey(b)
	if err != nil {
		return "", fmt.Errorf("unmarshal libp2p/secp256k1 pubkey: %w", err)
	}
	raw, err := pub.Raw()
	if err != nil {
		return "", fmt.Errorf("extract secp256k1 raw pubkey: %w", err)
	}
	if len(raw) != 33 {
		return "", fmt.Errorf("unexpected raw pubkey length: %d (expect 33)", len(raw))
	}
	return strings.ToLower(hex.EncodeToString(raw)), nil
}

// NormalizeClientIDStrict 统一 client_id 到 secp256k1 压缩公钥 hex（小写）。
// 说明：输入支持历史 marshal 格式与标准 compressed 格式；非法输入返回错误。
func NormalizeClientIDStrict(clientID string) (string, error) {
	return Libp2pMarshalPubHexToSecpCompressedHex(clientID)
}

// NormalizeClientIDLoose 尝试规范化 client_id；无法解析时仅做 lower+trim。
// 说明：该函数用于历史数据迁移与兼容查询，避免旧测试夹具（如 client_a）被硬失败。
func NormalizeClientIDLoose(clientID string) string {
	v := strings.ToLower(strings.TrimSpace(clientID))
	if v == "" {
		return ""
	}
	if norm, err := NormalizeClientIDStrict(v); err == nil {
		return norm
	}
	return v
}

// ClientIDAliasesForQuery 返回查询别名集合（canonical + legacy marshal + 输入本身）。
// 说明：用于过渡期兼容旧数据，迁移完成后理论上只命中 canonical。
func ClientIDAliasesForQuery(clientID string) []string {
	out := make([]string, 0, 3)
	seen := map[string]struct{}{}
	add := func(v string) {
		v = strings.ToLower(strings.TrimSpace(v))
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	add(clientID)
	canon, err := NormalizeClientIDStrict(clientID)
	if err != nil {
		return out
	}
	add(canon)
	if legacy, err := MarshalClientIDHexFromCanonical(canon); err == nil {
		add(legacy)
	}
	return out
}

// MarshalClientIDHexFromCanonical 将 canonical compressed pubkey 还原为历史 marshal hex。
func MarshalClientIDHexFromCanonical(canonicalHex string) (string, error) {
	canonicalHex = strings.TrimSpace(canonicalHex)
	b, err := hex.DecodeString(canonicalHex)
	if err != nil {
		return "", err
	}
	pub, err := crypto.UnmarshalSecp256k1PublicKey(b)
	if err != nil {
		return "", err
	}
	raw, err := crypto.MarshalPublicKey(pub)
	if err != nil {
		return "", err
	}
	return strings.ToLower(hex.EncodeToString(raw)), nil
}
