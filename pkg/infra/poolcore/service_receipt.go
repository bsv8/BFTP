package dual2of2

import (
	"fmt"
	"strings"
	"time"

	"github.com/bsv8/BFTP/pkg/infra/payflow"
)

// BuildSignedServiceReceipt 基于一次已受理扣费结果生成链下业务回执。
// 设计说明：
// - 回执必须绑定 accepted_charge_hash，避免把别的业务结果嫁接到本次收费；
// - payloadBytes 由业务层自己定义，只要求 client/server 使用同一份规范化编码。
func BuildSignedServiceReceipt(serverPrivHex string, isMainnet bool, clientPubkeyHex string, payOut PayConfirmResp, serviceType string, resultCode string, payloadBytes []byte) ([]byte, error) {
	if strings.TrimSpace(serverPrivHex) == "" {
		return nil, fmt.Errorf("server private key required")
	}
	state, err := proof.UnmarshalProofState(payOut.ProofStatePayload)
	if err != nil {
		return nil, fmt.Errorf("decode proof state payload: %w", err)
	}
	actor, err := BuildActor("service_receipt", strings.TrimSpace(serverPrivHex), isMainnet)
	if err != nil {
		return nil, err
	}
	receipt, err := proof.SignServiceReceipt(proof.ServiceReceipt{
		ServiceType:        strings.TrimSpace(serviceType),
		GatewayPubkeyHex:   strings.ToLower(strings.TrimSpace(actor.PubHex)),
		ClientPubkeyHex:    NormalizeClientIDLoose(clientPubkeyHex),
		SpendTxID:          strings.TrimSpace(state.SpendTxID),
		SequenceNumber:     state.SequenceNumber,
		AcceptedChargeHash: strings.ToLower(strings.TrimSpace(state.LastAcceptedChargeHash)),
		ResultCode:         strings.TrimSpace(resultCode),
		ResultPayloadHash:  proof.HashPayloadBytes(payloadBytes),
		CompletedAtUnix:    time.Now().Unix(),
	}, actor.PrivKey)
	if err != nil {
		return nil, err
	}
	return proof.MarshalServiceReceipt(receipt)
}
