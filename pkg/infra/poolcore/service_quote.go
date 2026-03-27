package poolcore

import (
	"fmt"
	"strings"
	"time"

	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv8/BFTP/pkg/infra/payflow"
)

const (
	QuoteServiceTypeListenCycle              = "listen_cycle_fee"
	QuoteServiceTypeDemandPublish            = "demand_publish_fee"
	QuoteServiceTypeDemandPublishBatch       = "demand_publish_batch_fee"
	QuoteServiceTypeLiveDemandPublish        = "live_demand_publish_fee"
	QuoteServiceTypeNodeReachabilityAnnounce = "node_reachability_announce_fee"
	QuoteServiceTypeNodeReachabilityQuery    = "node_reachability_query_fee"

	ServiceOfferPricingModeFixedPrice       = "fixed_price"
	ServiceOfferPricingModeBudgetForService = "budget_for_service"

	defaultServiceQuoteTTLSeconds = uint32(30)
)

type ServiceQuoteBuildInput struct {
	Offer                      payflow.ServiceOffer
	ChargeReason               string
	ChargeAmountSatoshi        uint64
	GrantedServiceDeadlineUnix int64
	GrantedDurationSeconds     uint32
	QuoteTTLSeconds            uint32
}

func HashServiceParamsPayload(raw []byte) string {
	return payflow.HashPayloadBytes(raw)
}

func ParseAndVerifyServiceQuote(raw []byte, gatewayPub *ec.PublicKey) (payflow.ServiceQuote, string, error) {
	quote, err := payflow.UnmarshalServiceQuote(raw)
	if err != nil {
		return payflow.ServiceQuote{}, "", err
	}
	if err := payflow.VerifyServiceQuoteSignature(quote, gatewayPub); err != nil {
		return payflow.ServiceQuote{}, "", err
	}
	hash, err := payflow.HashServiceQuote(quote)
	if err != nil {
		return payflow.ServiceQuote{}, "", err
	}
	return quote, hash, nil
}

func ValidateServiceQuoteBinding(quote payflow.ServiceQuote, expectedGatewayPubkeyHex string, expectedClientID string, expectedSpendTxID string, expectedServiceType string, serviceParamsPayload []byte, nowUnix int64) error {
	if expectedGatewayPubkeyHex != "" && !strings.EqualFold(strings.TrimSpace(quote.GatewayPubkeyHex), strings.TrimSpace(expectedGatewayPubkeyHex)) {
		return fmt.Errorf("service quote gateway_pubkey_hex mismatch")
	}
	if expectedClientID != "" && !strings.EqualFold(strings.TrimSpace(quote.ClientPubkeyHex), NormalizeClientIDLoose(expectedClientID)) {
		return fmt.Errorf("service quote client_pubkey_hex mismatch")
	}
	if expectedSpendTxID != "" && !strings.EqualFold(strings.TrimSpace(quote.SpendTxID), strings.TrimSpace(expectedSpendTxID)) {
		return fmt.Errorf("service quote spend_txid mismatch")
	}
	if expectedServiceType != "" && !strings.EqualFold(strings.TrimSpace(quote.ServiceType), strings.TrimSpace(expectedServiceType)) {
		return fmt.Errorf("service quote service_type mismatch")
	}
	if serviceParamsPayload != nil {
		if !strings.EqualFold(strings.TrimSpace(quote.ServiceParamsHash), HashServiceParamsPayload(serviceParamsPayload)) {
			return fmt.Errorf("service quote service_params_hash mismatch")
		}
	}
	if nowUnix > 0 && quote.QuoteExpiresAtUnix > 0 && nowUnix > quote.QuoteExpiresAtUnix {
		return fmt.Errorf("service quote expired")
	}
	return nil
}

func NormalizeServiceOfferPricingMode(raw string) string {
	mode := strings.TrimSpace(raw)
	if mode == "" {
		return ""
	}
	switch mode {
	case ServiceOfferPricingModeFixedPrice, ServiceOfferPricingModeBudgetForService:
		return mode
	default:
		return mode
	}
}

func (s *GatewayService) BuildServiceQuote(input ServiceQuoteBuildInput) (payflow.ServiceQuote, []byte, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s == nil || s.DB == nil {
		return payflow.ServiceQuote{}, nil, "", fmt.Errorf("db not initialized")
	}
	offer := input.Offer.Normalize()
	if err := offer.Validate(); err != nil {
		return payflow.ServiceQuote{}, nil, "", err
	}
	if input.ChargeAmountSatoshi == 0 {
		return payflow.ServiceQuote{}, nil, "", fmt.Errorf("charge_amount_satoshi required")
	}
	row, found, err := LoadSessionBySpendTxID(s.DB, offer.SpendTxID)
	if err != nil {
		return payflow.ServiceQuote{}, nil, "", err
	}
	if !found {
		return payflow.ServiceQuote{}, nil, "", fmt.Errorf("fee pool session not found")
	}
	if !strings.EqualFold(strings.TrimSpace(row.ClientID), strings.TrimSpace(offer.ClientPubkeyHex)) {
		return payflow.ServiceQuote{}, nil, "", fmt.Errorf("service offer client_pubkey_hex mismatch")
	}
	if err := ensureActivePoolGate(row); err != nil {
		return payflow.ServiceQuote{}, nil, "", err
	}
	if !s.isUniqueActiveSession(row.ClientID, row.SpendTxID) {
		return payflow.ServiceQuote{}, nil, "", fmt.Errorf("client has another active pool")
	}
	_, phase, payability, _, err := s.queryPhase(row)
	if err != nil {
		return payflow.ServiceQuote{}, nil, "", err
	}
	if payability != payabilityPayable {
		if phase == phaseNearExpiry {
			return payflow.ServiceQuote{}, nil, "", fmt.Errorf("fee pool near expiry")
		}
		return payflow.ServiceQuote{}, nil, "", fmt.Errorf("fee pool should submit")
	}
	nextServerAmount := row.ServerAmountSat + input.ChargeAmountSatoshi
	if nextServerAmount+row.SpendTxFeeSat > row.PoolAmountSat {
		return payflow.ServiceQuote{}, nil, "", fmt.Errorf("pool cannot cover charge")
	}
	serverActor, err := BuildActor("gateway", strings.TrimSpace(s.ServerPrivHex), s.IsMainnet)
	if err != nil {
		return payflow.ServiceQuote{}, nil, "", err
	}
	if !strings.EqualFold(strings.TrimSpace(offer.GatewayPubkeyHex), strings.TrimSpace(serverActor.PubHex)) {
		return payflow.ServiceQuote{}, nil, "", fmt.Errorf("service offer gateway_pubkey_hex mismatch")
	}
	offer.PricingMode = NormalizeServiceOfferPricingMode(offer.PricingMode)
	switch offer.PricingMode {
	case ServiceOfferPricingModeFixedPrice:
	case ServiceOfferPricingModeBudgetForService:
		if offer.ProposedPaymentSatoshi == 0 {
			return payflow.ServiceQuote{}, nil, "", fmt.Errorf("service offer proposed_payment_satoshi required")
		}
	default:
		return payflow.ServiceQuote{}, nil, "", fmt.Errorf("service offer pricing_mode unsupported")
	}
	offerHash, err := payflow.HashServiceOffer(offer)
	if err != nil {
		return payflow.ServiceQuote{}, nil, "", err
	}
	nowUnix := time.Now().Unix()
	ttlSeconds := input.QuoteTTLSeconds
	if ttlSeconds == 0 {
		ttlSeconds = defaultServiceQuoteTTLSeconds
	}
	signed, err := payflow.SignServiceQuote(payflow.ServiceQuote{
		OfferHash:                  offerHash,
		Domain:                     offer.Domain,
		ServiceType:                offer.ServiceType,
		ChargeReason:               strings.TrimSpace(input.ChargeReason),
		Target:                     offer.Target,
		GatewayPubkeyHex:           strings.ToLower(strings.TrimSpace(serverActor.PubHex)),
		ClientPubkeyHex:            strings.ToLower(strings.TrimSpace(row.ClientID)),
		SpendTxID:                  strings.TrimSpace(row.SpendTxID),
		ServiceParamsHash:          offer.ServiceParamsHash,
		SequenceNumber:             row.Sequence + 1,
		ServerAmountBefore:         row.ServerAmountSat,
		ChargeAmountSatoshi:        input.ChargeAmountSatoshi,
		ServerAmountAfter:          nextServerAmount,
		GrantedServiceDeadlineUnix: input.GrantedServiceDeadlineUnix,
		GrantedDurationSeconds:     input.GrantedDurationSeconds,
		QuoteExpiresAtUnix:         nowUnix + int64(ttlSeconds),
		IssuedAtUnix:               nowUnix,
	}, serverActor.PrivKey)
	if err != nil {
		return payflow.ServiceQuote{}, nil, "", err
	}
	raw, err := payflow.MarshalServiceQuote(signed)
	if err != nil {
		return payflow.ServiceQuote{}, nil, "", err
	}
	status := "accepted"
	if offer.PricingMode == ServiceOfferPricingModeBudgetForService && offer.ProposedPaymentSatoshi != input.ChargeAmountSatoshi {
		status = "countered"
	}
	return signed, raw, status, nil
}

func ListenOfferBudgetToDurationSeconds(proposedPayment uint64, minimumPayment uint64, minimumDurationSeconds uint32) (uint64, error) {
	const maxGrantedDurationSeconds = uint64(^uint32(0))
	if minimumPayment == 0 {
		return 0, fmt.Errorf("minimum payment required")
	}
	if minimumDurationSeconds == 0 {
		return 0, fmt.Errorf("minimum duration required")
	}
	if proposedPayment == 0 {
		return 0, fmt.Errorf("proposed payment required")
	}
	product := proposedPayment * uint64(minimumDurationSeconds)
	if proposedPayment != 0 && product/proposedPayment != uint64(minimumDurationSeconds) {
		return maxGrantedDurationSeconds, nil
	}
	duration := product / minimumPayment
	if duration == 0 {
		duration = 1
	}
	if duration > maxGrantedDurationSeconds {
		return maxGrantedDurationSeconds, nil
	}
	return duration, nil
}
