package dual2of2

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// Libp2pMarshalPubHexToSecpCompressedHex 把 libp2p MarshalPublicKey(hex) 转成 secp256k1 压缩公钥(hex，33字节)。
// 这是把现有 p2prpc 的 client_id/sender_pubkey 语义，桥接到 KeymasterMultisigPool 的 BSV 公钥语义上。
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
	if err != nil {
		return "", fmt.Errorf("unmarshal libp2p pubkey: %w", err)
	}
	raw, err := pub.Raw()
	if err != nil {
		return "", fmt.Errorf("extract raw pubkey: %w", err)
	}
	// KeymasterMultisigPool 需要 secp256k1 压缩公钥（33字节）。
	if len(raw) != 33 {
		return "", fmt.Errorf("unexpected raw pubkey length: %d (expect 33)", len(raw))
	}
	return strings.ToLower(hex.EncodeToString(raw)), nil
}

