package proof

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func UnmarshalServiceOffer(raw []byte) (ServiceOffer, error) {
	var parts []json.RawMessage
	if err := json.Unmarshal(raw, &parts); err != nil {
		return ServiceOffer{}, fmt.Errorf("decode service offer: %w", err)
	}
	if len(parts) != 11 {
		return ServiceOffer{}, fmt.Errorf("service offer fields mismatch")
	}
	var out ServiceOffer
	if err := json.Unmarshal(parts[0], &out.Version); err != nil {
		return ServiceOffer{}, err
	}
	if err := json.Unmarshal(parts[1], &out.Domain); err != nil {
		return ServiceOffer{}, err
	}
	if err := json.Unmarshal(parts[2], &out.ServiceType); err != nil {
		return ServiceOffer{}, err
	}
	if err := json.Unmarshal(parts[3], &out.Target); err != nil {
		return ServiceOffer{}, err
	}
	if err := json.Unmarshal(parts[4], &out.GatewayPubkeyHex); err != nil {
		return ServiceOffer{}, err
	}
	if err := json.Unmarshal(parts[5], &out.ClientPubkeyHex); err != nil {
		return ServiceOffer{}, err
	}
	if err := json.Unmarshal(parts[6], &out.SpendTxID); err != nil {
		return ServiceOffer{}, err
	}
	if err := json.Unmarshal(parts[7], &out.ServiceParamsHash); err != nil {
		return ServiceOffer{}, err
	}
	if err := json.Unmarshal(parts[8], &out.PricingMode); err != nil {
		return ServiceOffer{}, err
	}
	if err := json.Unmarshal(parts[9], &out.ProposedPaymentSatoshi); err != nil {
		return ServiceOffer{}, err
	}
	if err := json.Unmarshal(parts[10], &out.CreatedAtUnix); err != nil {
		return ServiceOffer{}, err
	}
	return out.Normalize(), out.Validate()
}

func UnmarshalServiceQuote(raw []byte) (ServiceQuote, error) {
	fields, signatureHex, err := unmarshalSignedArrayEnvelope(raw)
	if err != nil {
		return ServiceQuote{}, fmt.Errorf("decode service quote: %w", err)
	}
	if len(fields) != 18 {
		return ServiceQuote{}, fmt.Errorf("service quote fields mismatch")
	}
	var out ServiceQuote
	if err := json.Unmarshal(fields[0], &out.Version); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[1], &out.OfferHash); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[2], &out.Domain); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[3], &out.ServiceType); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[4], &out.ChargeReason); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[5], &out.Target); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[6], &out.GatewayPubkeyHex); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[7], &out.ClientPubkeyHex); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[8], &out.SpendTxID); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[9], &out.ServiceParamsHash); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[10], &out.SequenceNumber); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[11], &out.ServerAmountBefore); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[12], &out.ChargeAmountSatoshi); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[13], &out.ServerAmountAfter); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[14], &out.GrantedServiceDeadlineUnix); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[15], &out.GrantedDurationSeconds); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[16], &out.QuoteExpiresAtUnix); err != nil {
		return ServiceQuote{}, err
	}
	if err := json.Unmarshal(fields[17], &out.IssuedAtUnix); err != nil {
		return ServiceQuote{}, err
	}
	out.GatewaySignatureHex = signatureHex
	return out.Normalize(), out.Validate()
}

func UnmarshalIntent(raw []byte) (ChargeIntent, error) {
	var parts []json.RawMessage
	if err := json.Unmarshal(raw, &parts); err != nil {
		return ChargeIntent{}, fmt.Errorf("decode charge intent: %w", err)
	}
	if len(parts) != 13 {
		return ChargeIntent{}, fmt.Errorf("charge intent fields mismatch")
	}
	var out ChargeIntent
	if err := json.Unmarshal(parts[0], &out.Version); err != nil {
		return ChargeIntent{}, err
	}
	if err := json.Unmarshal(parts[1], &out.Domain); err != nil {
		return ChargeIntent{}, err
	}
	if err := json.Unmarshal(parts[2], &out.Target); err != nil {
		return ChargeIntent{}, err
	}
	if err := json.Unmarshal(parts[3], &out.GatewayPubkeyHex); err != nil {
		return ChargeIntent{}, err
	}
	if err := json.Unmarshal(parts[4], &out.ClientPubkeyHex); err != nil {
		return ChargeIntent{}, err
	}
	if err := json.Unmarshal(parts[5], &out.SpendTxID); err != nil {
		return ChargeIntent{}, err
	}
	if err := json.Unmarshal(parts[6], &out.GatewayQuoteHash); err != nil {
		return ChargeIntent{}, err
	}
	if err := json.Unmarshal(parts[7], &out.ChargeReason); err != nil {
		return ChargeIntent{}, err
	}
	if err := json.Unmarshal(parts[8], &out.ChargeAmountSatoshi); err != nil {
		return ChargeIntent{}, err
	}
	if err := json.Unmarshal(parts[9], &out.SequenceNumber); err != nil {
		return ChargeIntent{}, err
	}
	if err := json.Unmarshal(parts[10], &out.ServerAmountBefore); err != nil {
		return ChargeIntent{}, err
	}
	if err := json.Unmarshal(parts[11], &out.ServerAmountAfter); err != nil {
		return ChargeIntent{}, err
	}
	if err := json.Unmarshal(parts[12], &out.ServiceDeadlineUnix); err != nil {
		return ChargeIntent{}, err
	}
	return out.Normalize(), out.Validate()
}

func UnmarshalClientCommit(raw []byte) (ClientCommit, error) {
	var parts []json.RawMessage
	if err := json.Unmarshal(raw, &parts); err != nil {
		return ClientCommit{}, fmt.Errorf("decode client commit: %w", err)
	}
	return unmarshalClientCommitFields(parts)
}

func unmarshalClientCommitFields(parts []json.RawMessage) (ClientCommit, error) {
	if len(parts) != 10 {
		return ClientCommit{}, fmt.Errorf("client commit fields mismatch")
	}
	var out ClientCommit
	if err := json.Unmarshal(parts[0], &out.Version); err != nil {
		return ClientCommit{}, err
	}
	if err := json.Unmarshal(parts[1], &out.IntentHash); err != nil {
		return ClientCommit{}, err
	}
	if err := json.Unmarshal(parts[2], &out.ClientPubkeyHex); err != nil {
		return ClientCommit{}, err
	}
	if err := json.Unmarshal(parts[3], &out.SpendTxID); err != nil {
		return ClientCommit{}, err
	}
	if err := json.Unmarshal(parts[4], &out.SequenceNumber); err != nil {
		return ClientCommit{}, err
	}
	if err := json.Unmarshal(parts[5], &out.ServerAmountBefore); err != nil {
		return ClientCommit{}, err
	}
	if err := json.Unmarshal(parts[6], &out.ChargeAmountSatoshi); err != nil {
		return ClientCommit{}, err
	}
	if err := json.Unmarshal(parts[7], &out.ServerAmountAfter); err != nil {
		return ClientCommit{}, err
	}
	if err := json.Unmarshal(parts[8], &out.UpdateTemplateHash); err != nil {
		return ClientCommit{}, err
	}
	if err := json.Unmarshal(parts[9], &out.CreatedAtUnix); err != nil {
		return ClientCommit{}, err
	}
	return out.Normalize(), out.Validate()
}

func UnmarshalSignedClientCommit(raw []byte) (ClientCommit, []byte, error) {
	fields, signatureHex, err := unmarshalSignedArrayEnvelope(raw)
	if err != nil {
		return ClientCommit{}, nil, fmt.Errorf("decode signed client commit: %w", err)
	}
	commit, err := unmarshalClientCommitFields(fields)
	if err != nil {
		return ClientCommit{}, nil, err
	}
	sig, err := hex.DecodeString(normalizeHex(signatureHex))
	if err != nil {
		return ClientCommit{}, nil, fmt.Errorf("decode client commit signature: %w", err)
	}
	if len(sig) == 0 {
		return ClientCommit{}, nil, fmt.Errorf("client commit signature required")
	}
	return commit, sig, nil
}

func UnmarshalAcceptedCharge(raw []byte) (AcceptedCharge, error) {
	var parts []json.RawMessage
	if err := json.Unmarshal(raw, &parts); err != nil {
		return AcceptedCharge{}, fmt.Errorf("decode accepted charge: %w", err)
	}
	if len(parts) != 10 {
		return AcceptedCharge{}, fmt.Errorf("accepted charge fields mismatch")
	}
	var out AcceptedCharge
	if err := json.Unmarshal(parts[0], &out.Version); err != nil {
		return AcceptedCharge{}, err
	}
	if err := json.Unmarshal(parts[1], &out.IntentHash); err != nil {
		return AcceptedCharge{}, err
	}
	if err := json.Unmarshal(parts[2], &out.ClientCommitHash); err != nil {
		return AcceptedCharge{}, err
	}
	if err := json.Unmarshal(parts[3], &out.SpendTxID); err != nil {
		return AcceptedCharge{}, err
	}
	if err := json.Unmarshal(parts[4], &out.SequenceNumber); err != nil {
		return AcceptedCharge{}, err
	}
	if err := json.Unmarshal(parts[5], &out.ServerAmountBefore); err != nil {
		return AcceptedCharge{}, err
	}
	if err := json.Unmarshal(parts[6], &out.ChargeAmountSatoshi); err != nil {
		return AcceptedCharge{}, err
	}
	if err := json.Unmarshal(parts[7], &out.ServerAmountAfter); err != nil {
		return AcceptedCharge{}, err
	}
	if err := json.Unmarshal(parts[8], &out.ServiceDeadlineUnix); err != nil {
		return AcceptedCharge{}, err
	}
	if err := json.Unmarshal(parts[9], &out.PrevAcceptedHash); err != nil {
		return AcceptedCharge{}, err
	}
	return out.Normalize(), out.Validate()
}

func UnmarshalProofState(raw []byte) (ProofState, error) {
	var parts []json.RawMessage
	if err := json.Unmarshal(raw, &parts); err != nil {
		return ProofState{}, fmt.Errorf("decode proof state: %w", err)
	}
	if len(parts) != 7 {
		return ProofState{}, fmt.Errorf("proof state fields mismatch")
	}
	var out ProofState
	if err := json.Unmarshal(parts[0], &out.Version); err != nil {
		return ProofState{}, err
	}
	if err := json.Unmarshal(parts[1], &out.SpendTxID); err != nil {
		return ProofState{}, err
	}
	if err := json.Unmarshal(parts[2], &out.SequenceNumber); err != nil {
		return ProofState{}, err
	}
	if err := json.Unmarshal(parts[3], &out.ServerAmountSatoshi); err != nil {
		return ProofState{}, err
	}
	if err := json.Unmarshal(parts[4], &out.AcceptedTipHash); err != nil {
		return ProofState{}, err
	}
	if err := json.Unmarshal(parts[5], &out.LastAcceptedChargeHash); err != nil {
		return ProofState{}, err
	}
	if err := json.Unmarshal(parts[6], &out.ServiceDeadlineUnix); err != nil {
		return ProofState{}, err
	}
	return out.Normalize(), out.Validate()
}

func UnmarshalServiceReceipt(raw []byte) (ServiceReceipt, error) {
	fields, signatureHex, err := unmarshalSignedArrayEnvelope(raw)
	if err != nil {
		return ServiceReceipt{}, fmt.Errorf("decode service receipt: %w", err)
	}
	if len(fields) != 10 {
		return ServiceReceipt{}, fmt.Errorf("service receipt fields mismatch")
	}
	var out ServiceReceipt
	if err := json.Unmarshal(fields[0], &out.Version); err != nil {
		return ServiceReceipt{}, err
	}
	if err := json.Unmarshal(fields[1], &out.ServiceType); err != nil {
		return ServiceReceipt{}, err
	}
	if err := json.Unmarshal(fields[2], &out.GatewayPubkeyHex); err != nil {
		return ServiceReceipt{}, err
	}
	if err := json.Unmarshal(fields[3], &out.ClientPubkeyHex); err != nil {
		return ServiceReceipt{}, err
	}
	if err := json.Unmarshal(fields[4], &out.SpendTxID); err != nil {
		return ServiceReceipt{}, err
	}
	if err := json.Unmarshal(fields[5], &out.SequenceNumber); err != nil {
		return ServiceReceipt{}, err
	}
	if err := json.Unmarshal(fields[6], &out.AcceptedChargeHash); err != nil {
		return ServiceReceipt{}, err
	}
	if err := json.Unmarshal(fields[7], &out.ResultCode); err != nil {
		return ServiceReceipt{}, err
	}
	if err := json.Unmarshal(fields[8], &out.ResultPayloadHash); err != nil {
		return ServiceReceipt{}, err
	}
	if err := json.Unmarshal(fields[9], &out.CompletedAtUnix); err != nil {
		return ServiceReceipt{}, err
	}
	out.GatewaySignatureHex = signatureHex
	return out.Normalize(), out.Validate()
}
