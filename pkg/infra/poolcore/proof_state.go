package dual2of2

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv8/BFTP/pkg/infra/payflow"
)

func buildPayConfirmProofPayload(row GatewaySessionRow, req PayConfirmReq, updatedTxHex string) ([]byte, []byte, error) {
	if len(req.ProofIntent) == 0 && len(req.SignedProofCommit) == 0 {
		return nil, nil, nil
	}
	if len(req.ProofIntent) == 0 || len(req.SignedProofCommit) == 0 {
		return nil, nil, fmt.Errorf("proof intent and signed_proof_commit must be provided together")
	}
	intent, err := proof.UnmarshalIntent(req.ProofIntent)
	if err != nil {
		return nil, nil, err
	}
	_, quoteHash, err := validateServiceQuoteAgainstPay(row, req)
	if err != nil {
		return nil, nil, err
	}
	if err := validateProofIntentAgainstPay(row, req, intent, quoteHash); err != nil {
		return nil, nil, err
	}
	clientPub, err := ec.PublicKeyFromString(strings.TrimSpace(row.ClientBSVCompressedPubHex))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid stored client pubkey: %w", err)
	}
	commit, err := proof.VerifySignedClientCommit(req.SignedProofCommit, clientPub)
	if err != nil {
		return nil, nil, err
	}
	if err := validateProofCommitAgainstPay(row, req, updatedTxHex, intent, commit); err != nil {
		return nil, nil, err
	}
	prevState, found, err := proof.ExtractProofStateFromTxHex(row.CurrentTxHex)
	if err != nil {
		return nil, nil, err
	}
	if found && !strings.EqualFold(prevState.SpendTxID, row.SpendTxID) {
		return nil, nil, fmt.Errorf("previous proof state spend_txid mismatch")
	}
	intentHash, err := proof.HashIntent(intent)
	if err != nil {
		return nil, nil, err
	}
	commitHash, err := proof.HashClientCommit(commit)
	if err != nil {
		return nil, nil, err
	}
	accepted := proof.AcceptedCharge{
		IntentHash:          intentHash,
		ClientCommitHash:    commitHash,
		SpendTxID:           row.SpendTxID,
		SequenceNumber:      req.SequenceNumber,
		ServerAmountBefore:  row.ServerAmountSat,
		ChargeAmountSatoshi: req.ChargeAmountSatoshi,
		ServerAmountAfter:   req.ServerAmount,
		ServiceDeadlineUnix: intent.ServiceDeadlineUnix,
		PrevAcceptedHash:    prevState.AcceptedTipHash,
	}
	state, err := proof.BuildNextProofState(prevState, accepted)
	if err != nil {
		return nil, nil, err
	}
	payload, err := proof.MarshalProofState(state)
	if err != nil {
		return nil, nil, err
	}
	acceptedRaw, err := proof.MarshalAcceptedCharge(accepted)
	if err != nil {
		return nil, nil, err
	}
	return payload, acceptedRaw, nil
}

func validateServiceQuoteAgainstPay(row GatewaySessionRow, req PayConfirmReq) (proof.ServiceQuote, string, error) {
	if len(req.ServiceQuote) == 0 {
		return proof.ServiceQuote{}, "", nil
	}
	gatewayPub, err := ec.PublicKeyFromString(strings.TrimSpace(row.ServerBSVCompressedPubHex))
	if err != nil {
		return proof.ServiceQuote{}, "", fmt.Errorf("invalid stored gateway pubkey: %w", err)
	}
	quote, quoteHash, err := ParseAndVerifyServiceQuote(req.ServiceQuote, gatewayPub)
	if err != nil {
		return proof.ServiceQuote{}, "", err
	}
	if err := ValidateServiceQuoteBinding(quote, row.ServerBSVCompressedPubHex, row.ClientBSVCompressedPubHex, row.SpendTxID, "", nil, time.Now().Unix()); err != nil {
		return proof.ServiceQuote{}, "", err
	}
	if !strings.EqualFold(strings.TrimSpace(quote.ChargeReason), strings.TrimSpace(req.ChargeReason)) {
		return proof.ServiceQuote{}, "", fmt.Errorf("service quote charge_reason mismatch")
	}
	if quote.SequenceNumber != req.SequenceNumber {
		return proof.ServiceQuote{}, "", fmt.Errorf("service quote sequence_number mismatch")
	}
	if quote.ServerAmountBefore != row.ServerAmountSat {
		return proof.ServiceQuote{}, "", fmt.Errorf("service quote server_amount_before mismatch")
	}
	if quote.ChargeAmountSatoshi != req.ChargeAmountSatoshi {
		return proof.ServiceQuote{}, "", fmt.Errorf("service quote charge_amount mismatch")
	}
	if quote.ServerAmountAfter != req.ServerAmount {
		return proof.ServiceQuote{}, "", fmt.Errorf("service quote server_amount_after mismatch")
	}
	return quote, quoteHash, nil
}

func validateProofIntentAgainstPay(row GatewaySessionRow, req PayConfirmReq, intent proof.ChargeIntent, quoteHash string) error {
	if !strings.EqualFold(strings.TrimSpace(intent.SpendTxID), row.SpendTxID) {
		return fmt.Errorf("proof intent spend_txid mismatch")
	}
	if strings.TrimSpace(intent.GatewayPubkeyHex) != "" && !strings.EqualFold(strings.TrimSpace(intent.GatewayPubkeyHex), row.ServerBSVCompressedPubHex) {
		return fmt.Errorf("proof intent gateway_pubkey_hex mismatch")
	}
	if strings.TrimSpace(intent.ClientPubkeyHex) != "" && !strings.EqualFold(strings.TrimSpace(intent.ClientPubkeyHex), row.ClientBSVCompressedPubHex) {
		return fmt.Errorf("proof intent client_pubkey_hex mismatch")
	}
	if quoteHash != "" && !strings.EqualFold(strings.TrimSpace(intent.GatewayQuoteHash), strings.TrimSpace(quoteHash)) {
		return fmt.Errorf("proof intent gateway_quote_hash mismatch")
	}
	if !strings.EqualFold(strings.TrimSpace(intent.ChargeReason), strings.TrimSpace(req.ChargeReason)) {
		return fmt.Errorf("proof intent charge_reason mismatch")
	}
	if intent.ChargeAmountSatoshi != req.ChargeAmountSatoshi {
		return fmt.Errorf("proof intent charge_amount mismatch")
	}
	if intent.SequenceNumber != req.SequenceNumber {
		return fmt.Errorf("proof intent sequence_number mismatch")
	}
	if intent.ServerAmountBefore != row.ServerAmountSat {
		return fmt.Errorf("proof intent server_amount_before mismatch")
	}
	if intent.ServerAmountAfter != req.ServerAmount {
		return fmt.Errorf("proof intent server_amount_after mismatch")
	}
	return nil
}

func validateProofCommitAgainstPay(row GatewaySessionRow, req PayConfirmReq, updatedTxHex string, intent proof.ChargeIntent, commit proof.ClientCommit) error {
	intentHash, err := proof.HashIntent(intent)
	if err != nil {
		return err
	}
	if !strings.EqualFold(commit.IntentHash, intentHash) {
		return fmt.Errorf("proof commit intent_hash mismatch")
	}
	if !strings.EqualFold(strings.TrimSpace(commit.ClientPubkeyHex), row.ClientBSVCompressedPubHex) {
		return fmt.Errorf("proof commit client_pubkey_hex mismatch")
	}
	if !strings.EqualFold(strings.TrimSpace(commit.SpendTxID), row.SpendTxID) {
		return fmt.Errorf("proof commit spend_txid mismatch")
	}
	if commit.SequenceNumber != req.SequenceNumber {
		return fmt.Errorf("proof commit sequence_number mismatch")
	}
	if commit.ServerAmountBefore != row.ServerAmountSat {
		return fmt.Errorf("proof commit server_amount_before mismatch")
	}
	if commit.ChargeAmountSatoshi != req.ChargeAmountSatoshi {
		return fmt.Errorf("proof commit charge_amount mismatch")
	}
	if commit.ServerAmountAfter != req.ServerAmount {
		return fmt.Errorf("proof commit server_amount_after mismatch")
	}
	templateHash, err := proof.UpdateTemplateHashFromTxHex(updatedTxHex)
	if err != nil {
		return err
	}
	if !strings.EqualFold(commit.UpdateTemplateHash, templateHash) {
		return fmt.Errorf("proof commit update_template_hash mismatch")
	}
	return nil
}

func publicKeyHexFromEC(pub *ec.PublicKey) string {
	if pub == nil {
		return ""
	}
	return strings.ToLower(hex.EncodeToString(pub.Compressed()))
}
