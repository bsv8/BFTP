package payflow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	serviceOfferVersion   = "bsv8-service-offer-v3"
	serviceQuoteVersion   = "bsv8-service-quote-v3"
	intentVersion         = "bsv8-feepool-charge-intent-v1"
	clientCommitVersion   = "bsv8-feepool-client-commit-v1"
	acceptedChargeVersion = "bsv8-feepool-accepted-charge-v1"
	proofStateVersion     = "bsv8-feepool-proof-state-v1"
	serviceReceiptVersion = "bsv8-feepool-service-receipt-v2"
)

// ServiceOffer 描述 client 对某项服务的一次明确要约。
// 设计说明：
// - offer 只描述“我要买什么服务”，不再夹带费用池状态或支付方式；
// - request_params 保存业务原始请求，quote/pay 之后都通过 offer_hash 反查这一份原文；
// - OfferHash 是运行时派生值，不参与线格式，避免自引用。
type ServiceOffer struct {
	Version string

	ServiceType          string
	ServiceNodePubkeyHex string
	ClientPubkeyHex      string
	RequestParams        []byte
	CreatedAtUnix        int64
}

func (v ServiceOffer) Normalize() ServiceOffer {
	v.Version = normalizeOrDefault(v.Version, serviceOfferVersion)
	v.ServiceType = strings.TrimSpace(v.ServiceType)
	v.ServiceNodePubkeyHex = normalizeHex(v.ServiceNodePubkeyHex)
	v.ClientPubkeyHex = normalizeHex(v.ClientPubkeyHex)
	v.RequestParams = append([]byte(nil), v.RequestParams...)
	return v
}

func (v ServiceOffer) Validate() error {
	v = v.Normalize()
	if v.Version != serviceOfferVersion {
		return fmt.Errorf("service offer version mismatch")
	}
	if v.ServiceType == "" {
		return fmt.Errorf("service offer service_type required")
	}
	if v.ServiceNodePubkeyHex == "" {
		return fmt.Errorf("service offer service_node_pubkey_hex required")
	}
	if v.ClientPubkeyHex == "" {
		return fmt.Errorf("service offer client_pubkey_hex required")
	}
	if v.CreatedAtUnix <= 0 {
		return fmt.Errorf("service offer created_at_unix required")
	}
	return nil
}

func (v ServiceOffer) Array() []any {
	v = v.Normalize()
	return []any{
		v.Version,
		v.ServiceType,
		v.ServiceNodePubkeyHex,
		v.ClientPubkeyHex,
		v.RequestParams,
		v.CreatedAtUnix,
	}
}

// ServiceQuote 是 gateway 对某次要约给出的正式报价单。
// 设计说明：
// - quote 只保留最小成交条件，保证两种支付方式共用同一张报价；
// - 费用池状态、业务结果、payment_schemes 都不进入签名；
// - 线格式固定为 [[字段...], 签名]，便于直接进入 OP_RETURN。
type ServiceQuote struct {
	Version string

	OfferHash           string
	ChargeAmountSatoshi uint64
	ExpiresAtUnix       int64

	SignatureHex string
}

func (v ServiceQuote) Normalize() ServiceQuote {
	v.Version = normalizeOrDefault(v.Version, serviceQuoteVersion)
	v.OfferHash = normalizeHex(v.OfferHash)
	v.SignatureHex = normalizeHex(v.SignatureHex)
	return v
}

func (v ServiceQuote) ValidateUnsigned() error {
	v = v.Normalize()
	if v.Version != serviceQuoteVersion {
		return fmt.Errorf("service quote version mismatch")
	}
	if v.OfferHash == "" {
		return fmt.Errorf("service quote offer_hash required")
	}
	if v.ChargeAmountSatoshi == 0 {
		return fmt.Errorf("service quote charge_amount required")
	}
	if v.ExpiresAtUnix <= 0 {
		return fmt.Errorf("service quote expires_at_unix required")
	}
	return nil
}

func (v ServiceQuote) Validate() error {
	if err := v.ValidateUnsigned(); err != nil {
		return err
	}
	if v.Normalize().SignatureHex == "" {
		return fmt.Errorf("service quote signature required")
	}
	return nil
}

func (v ServiceQuote) UnsignedArray() []any {
	v = v.Normalize()
	return []any{
		v.Version,
		v.OfferHash,
		v.ChargeAmountSatoshi,
		v.ExpiresAtUnix,
	}
}

func (v ServiceQuote) Array() []any {
	v = v.Normalize()
	return []any{v.UnsignedArray(), v.SignatureHex}
}

// ChargeIntent 描述“一次收费意图”的最小公共语义。
// 设计说明：
// - 不把它命名成 quote，是为了让库能被 domain/gateway/未来别的收费子应用共用；
// - 现在 intent 必须引用 gateway_quote_hash，避免 client 脱离正式报价后再“自行发明”收费语义。
type ChargeIntent struct {
	Version string

	Domain           string
	Target           string
	GatewayPubkeyHex string
	ClientPubkeyHex  string
	SpendTxID        string
	GatewayQuoteHash string

	ChargeReason        string
	ChargeAmountSatoshi uint64

	SequenceNumber     uint32
	ServerAmountBefore uint64
	ServerAmountAfter  uint64

	ServiceDeadlineUnix int64
}

func (v ChargeIntent) Normalize() ChargeIntent {
	v.Version = normalizeOrDefault(v.Version, intentVersion)
	v.Domain = strings.TrimSpace(v.Domain)
	v.Target = strings.TrimSpace(v.Target)
	v.GatewayPubkeyHex = normalizeHex(v.GatewayPubkeyHex)
	v.ClientPubkeyHex = normalizeHex(v.ClientPubkeyHex)
	v.SpendTxID = strings.TrimSpace(v.SpendTxID)
	v.GatewayQuoteHash = normalizeHex(v.GatewayQuoteHash)
	v.ChargeReason = strings.TrimSpace(v.ChargeReason)
	return v
}

func (v ChargeIntent) Validate() error {
	v = v.Normalize()
	if v.Version != intentVersion {
		return fmt.Errorf("charge intent version mismatch")
	}
	if v.SpendTxID == "" {
		return fmt.Errorf("charge intent spend_txid required")
	}
	if v.GatewayQuoteHash == "" {
		return fmt.Errorf("charge intent gateway_quote_hash required")
	}
	if v.ChargeReason == "" {
		return fmt.Errorf("charge intent reason required")
	}
	if v.ChargeAmountSatoshi == 0 {
		return fmt.Errorf("charge intent amount required")
	}
	if v.ServerAmountAfter < v.ServerAmountBefore {
		return fmt.Errorf("charge intent server amount invalid")
	}
	if v.ServerAmountAfter-v.ServerAmountBefore != v.ChargeAmountSatoshi {
		return fmt.Errorf("charge intent amount delta mismatch")
	}
	return nil
}

func (v ChargeIntent) Array() []any {
	v = v.Normalize()
	return []any{
		v.Version,
		v.Domain,
		v.Target,
		v.GatewayPubkeyHex,
		v.ClientPubkeyHex,
		v.SpendTxID,
		v.GatewayQuoteHash,
		v.ChargeReason,
		v.ChargeAmountSatoshi,
		v.SequenceNumber,
		v.ServerAmountBefore,
		v.ServerAmountAfter,
		v.ServiceDeadlineUnix,
	}
}

// ClientCommit 是 client 对“某次收费意图 + 某个更新态模板”的明确绑定。
// 设计约束：
// - commit 只绑定 update template hash，不绑定最终 txid；
// - 这样 client 在 gateway 回签前也能先做不可抵赖承诺。
type ClientCommit struct {
	Version string

	IntentHash          string
	ClientPubkeyHex     string
	SpendTxID           string
	SequenceNumber      uint32
	ServerAmountBefore  uint64
	ChargeAmountSatoshi uint64
	ServerAmountAfter   uint64
	UpdateTemplateHash  string
	CreatedAtUnix       int64
}

func (v ClientCommit) Normalize() ClientCommit {
	v.Version = normalizeOrDefault(v.Version, clientCommitVersion)
	v.IntentHash = normalizeHex(v.IntentHash)
	v.ClientPubkeyHex = normalizeHex(v.ClientPubkeyHex)
	v.SpendTxID = strings.TrimSpace(v.SpendTxID)
	v.UpdateTemplateHash = normalizeHex(v.UpdateTemplateHash)
	return v
}

func (v ClientCommit) Validate() error {
	v = v.Normalize()
	if v.Version != clientCommitVersion {
		return fmt.Errorf("client commit version mismatch")
	}
	if v.IntentHash == "" {
		return fmt.Errorf("client commit intent_hash required")
	}
	if v.ClientPubkeyHex == "" {
		return fmt.Errorf("client commit client_pubkey_hex required")
	}
	if v.SpendTxID == "" {
		return fmt.Errorf("client commit spend_txid required")
	}
	if v.ChargeAmountSatoshi == 0 {
		return fmt.Errorf("client commit amount required")
	}
	if v.ServerAmountAfter < v.ServerAmountBefore {
		return fmt.Errorf("client commit server amount invalid")
	}
	if v.ServerAmountAfter-v.ServerAmountBefore != v.ChargeAmountSatoshi {
		return fmt.Errorf("client commit amount delta mismatch")
	}
	if v.UpdateTemplateHash == "" {
		return fmt.Errorf("client commit update_template_hash required")
	}
	return nil
}

func (v ClientCommit) Array() []any {
	v = v.Normalize()
	return []any{
		v.Version,
		v.IntentHash,
		v.ClientPubkeyHex,
		v.SpendTxID,
		v.SequenceNumber,
		v.ServerAmountBefore,
		v.ChargeAmountSatoshi,
		v.ServerAmountAfter,
		v.UpdateTemplateHash,
		v.CreatedAtUnix,
	}
}

// AcceptedCharge 是 gateway 真正受理收费后形成的链下主凭证。
// 最终 OP_RETURN 只锚这个对象的 hash 链头，不把完整原文全塞上链。
type AcceptedCharge struct {
	Version string

	IntentHash          string
	ClientCommitHash    string
	SpendTxID           string
	SequenceNumber      uint32
	ServerAmountBefore  uint64
	ChargeAmountSatoshi uint64
	ServerAmountAfter   uint64
	ServiceDeadlineUnix int64
	PrevAcceptedHash    string
}

func (v AcceptedCharge) Normalize() AcceptedCharge {
	v.Version = normalizeOrDefault(v.Version, acceptedChargeVersion)
	v.IntentHash = normalizeHex(v.IntentHash)
	v.ClientCommitHash = normalizeHex(v.ClientCommitHash)
	v.SpendTxID = strings.TrimSpace(v.SpendTxID)
	v.PrevAcceptedHash = normalizeHex(v.PrevAcceptedHash)
	return v
}

func (v AcceptedCharge) Validate() error {
	v = v.Normalize()
	if v.Version != acceptedChargeVersion {
		return fmt.Errorf("accepted charge version mismatch")
	}
	if v.IntentHash == "" {
		return fmt.Errorf("accepted charge intent_hash required")
	}
	if v.ClientCommitHash == "" {
		return fmt.Errorf("accepted charge client_commit_hash required")
	}
	if v.SpendTxID == "" {
		return fmt.Errorf("accepted charge spend_txid required")
	}
	if v.ChargeAmountSatoshi == 0 {
		return fmt.Errorf("accepted charge amount required")
	}
	if v.ServerAmountAfter < v.ServerAmountBefore {
		return fmt.Errorf("accepted charge server amount invalid")
	}
	if v.ServerAmountAfter-v.ServerAmountBefore != v.ChargeAmountSatoshi {
		return fmt.Errorf("accepted charge amount delta mismatch")
	}
	return nil
}

func (v AcceptedCharge) Array() []any {
	v = v.Normalize()
	return []any{
		v.Version,
		v.IntentHash,
		v.ClientCommitHash,
		v.SpendTxID,
		v.SequenceNumber,
		v.ServerAmountBefore,
		v.ChargeAmountSatoshi,
		v.ServerAmountAfter,
		v.ServiceDeadlineUnix,
		v.PrevAcceptedHash,
	}
}

// ProofState 是写入状态交易 OP_RETURN 的公开承诺。
// 设计说明：
// - 链上只承诺“已受理收费链头”，让最终 close tx 具备公开反推能力；
// - service receipt 暂时仍走链下举证，避免每次业务结果都反向牵连 MultisigPool 状态结构。
type ProofState struct {
	Version string

	SpendTxID           string
	SequenceNumber      uint32
	ServerAmountSatoshi uint64

	AcceptedTipHash        string
	LastAcceptedChargeHash string
	ServiceDeadlineUnix    int64
}

func (v ProofState) Normalize() ProofState {
	v.Version = normalizeOrDefault(v.Version, proofStateVersion)
	v.SpendTxID = strings.TrimSpace(v.SpendTxID)
	v.AcceptedTipHash = normalizeHex(v.AcceptedTipHash)
	v.LastAcceptedChargeHash = normalizeHex(v.LastAcceptedChargeHash)
	return v
}

func (v ProofState) Validate() error {
	v = v.Normalize()
	if v.Version != proofStateVersion {
		return fmt.Errorf("proof state version mismatch")
	}
	if v.SpendTxID == "" {
		return fmt.Errorf("proof state spend_txid required")
	}
	if v.SequenceNumber == 0 {
		return fmt.Errorf("proof state sequence_number required")
	}
	if v.AcceptedTipHash == "" {
		return fmt.Errorf("proof state accepted_tip_hash required")
	}
	if v.LastAcceptedChargeHash == "" {
		return fmt.Errorf("proof state last_accepted_charge_hash required")
	}
	if v.AcceptedTipHash != v.LastAcceptedChargeHash {
		return fmt.Errorf("proof state tip hash mismatch")
	}
	return nil
}

func (v ProofState) Array() []any {
	v = v.Normalize()
	return []any{
		v.Version,
		v.SpendTxID,
		v.SequenceNumber,
		v.ServerAmountSatoshi,
		v.AcceptedTipHash,
		v.LastAcceptedChargeHash,
		v.ServiceDeadlineUnix,
	}
}

// ServiceReceipt 是收费后的链下业务完成回执。
// 设计说明：
// - payment_receipt 只证明支付成立，service_receipt 只证明服务结果；
// - 回执绑定 offer_hash 与 result_hash，避免把别次业务结果嫁接到当前成交；
// - 线格式固定为 [[字段...], 签名]。
type ServiceReceipt struct {
	Version string

	OfferHash    string
	ServiceType  string
	ResultHash   string
	IssuedAtUnix int64

	SignatureHex string
}

func (v ServiceReceipt) Normalize() ServiceReceipt {
	v.Version = normalizeOrDefault(v.Version, serviceReceiptVersion)
	v.OfferHash = normalizeHex(v.OfferHash)
	v.ServiceType = strings.TrimSpace(v.ServiceType)
	v.ResultHash = normalizeHex(v.ResultHash)
	v.SignatureHex = normalizeHex(v.SignatureHex)
	return v
}

func (v ServiceReceipt) ValidateUnsigned() error {
	v = v.Normalize()
	if v.Version != serviceReceiptVersion {
		return fmt.Errorf("service receipt version mismatch")
	}
	if v.OfferHash == "" {
		return fmt.Errorf("service receipt offer_hash required")
	}
	if v.ServiceType == "" {
		return fmt.Errorf("service receipt service_type required")
	}
	if v.ResultHash == "" {
		return fmt.Errorf("service receipt result_hash required")
	}
	if v.IssuedAtUnix <= 0 {
		return fmt.Errorf("service receipt issued_at_unix required")
	}
	return nil
}

func (v ServiceReceipt) Validate() error {
	if err := v.ValidateUnsigned(); err != nil {
		return err
	}
	if v.Normalize().SignatureHex == "" {
		return fmt.Errorf("service receipt signature required")
	}
	return nil
}

func (v ServiceReceipt) UnsignedArray() []any {
	v = v.Normalize()
	return []any{
		v.Version,
		v.OfferHash,
		v.ServiceType,
		v.ResultHash,
		v.IssuedAtUnix,
	}
}

func (v ServiceReceipt) Array() []any {
	v = v.Normalize()
	return []any{v.UnsignedArray(), v.SignatureHex}
}

func normalizeOrDefault(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return strings.TrimSpace(v)
}

func normalizeHex(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func marshalArray(items []any) ([]byte, error) {
	raw, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func hashRawHex(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func HashIntent(v ChargeIntent) (string, error) {
	raw, err := marshalArray(v.Array())
	if err != nil {
		return "", err
	}
	return hashRawHex(raw), nil
}

func HashServiceOffer(v ServiceOffer) (string, error) {
	raw, err := marshalArray(v.Array())
	if err != nil {
		return "", err
	}
	return hashRawHex(raw), nil
}

func HashServiceQuote(v ServiceQuote) (string, error) {
	unsigned := v.Normalize()
	unsigned.SignatureHex = ""
	raw, err := marshalArray(unsigned.UnsignedArray())
	if err != nil {
		return "", err
	}
	return hashRawHex(raw), nil
}

func HashClientCommit(v ClientCommit) (string, error) {
	raw, err := marshalArray(v.Array())
	if err != nil {
		return "", err
	}
	return hashRawHex(raw), nil
}

func HashAcceptedCharge(v AcceptedCharge) (string, error) {
	raw, err := marshalArray(v.Array())
	if err != nil {
		return "", err
	}
	return hashRawHex(raw), nil
}

func HashPayloadBytes(raw []byte) string {
	return hashRawHex(raw)
}

func MarshalIntent(v ChargeIntent) ([]byte, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return marshalArray(v.Array())
}

func MarshalServiceOffer(v ServiceOffer) ([]byte, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return marshalArray(v.Array())
}

func MarshalServiceQuote(v ServiceQuote) ([]byte, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return marshalArray(v.Array())
}

func MarshalClientCommit(v ClientCommit) ([]byte, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return marshalArray(v.Array())
}

func MarshalSignedClientCommit(v ClientCommit, sig []byte) ([]byte, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	if len(sig) == 0 {
		return nil, fmt.Errorf("client commit signature required")
	}
	return marshalSignedArrayEnvelope(v.Array(), hex.EncodeToString(sig))
}

func MarshalAcceptedCharge(v AcceptedCharge) ([]byte, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return marshalArray(v.Array())
}

func MarshalProofState(v ProofState) ([]byte, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return marshalArray(v.Array())
}

func MarshalServiceReceipt(v ServiceReceipt) ([]byte, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return marshalArray(v.Array())
}
