package dual2of2

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bsv8/BFTP/pkg/obs"
	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	tx "github.com/bsv-blockchain/go-sdk/transaction"
	ce "github.com/bsv8/MultisigPool/pkg/dual_endpoint"
	"github.com/bsv8/MultisigPool/pkg/libs"
)

type GatewayParams struct {
	MinimumPoolAmountSatoshi uint64
	LockBlocks               uint32
	FeeRateSatPerByte        float64
	// PayRejectBeforeExpiryBlocks 控制“临近到期时拒绝继续 pay”的安全窗口（单位：区块）。
	// 取值为 0 时，会使用默认值 1。
	PayRejectBeforeExpiryBlocks uint32

	BillingCycleSeconds      uint32
	SingleCycleFeeSatoshi    uint64
	SinglePublishFeeSatoshi  uint64
	RenewNotifyBeforeSeconds uint32
}

type GatewayService struct {
	DB    *sql.DB
	Chain ChainClient

	ServerPrivHex string
	Params        GatewayParams

	// IsMainnet 控制地址派生与签名脚本等网络相关参数。
	// 说明：BSV 主网/测试网的地址格式不同；公私钥本身不变。
	IsMainnet bool

	mu sync.Mutex
}

const (
	feePoolLockTimeThreshold = uint32(500000000)
	defaultPayGuardBlocks    = uint32(1)
)

func (s *GatewayService) Info(clientID string) (InfoResp, error) {
	if strings.TrimSpace(clientID) == "" {
		return InfoResp{}, fmt.Errorf("client_id required")
	}
	return InfoResp{
		Status:                   "ok",
		MinimumPoolAmountSatoshi: s.Params.MinimumPoolAmountSatoshi,
		LockBlocks:               s.Params.LockBlocks,
		FeeRateSatPerByte:        s.Params.FeeRateSatPerByte,
		BillingCycleSeconds:      s.Params.BillingCycleSeconds,
		SingleCycleFeeSatoshi:    s.Params.SingleCycleFeeSatoshi,
		SinglePublishFeeSatoshi:  s.Params.SinglePublishFeeSatoshi,
		RenewNotifyBeforeSeconds: s.Params.RenewNotifyBeforeSeconds,
	}, nil
}

func (s *GatewayService) Create(req CreateReq) (CreateResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.DB == nil {
		return CreateResp{}, fmt.Errorf("db not initialized")
	}
	if strings.TrimSpace(req.ClientID) == "" {
		return CreateResp{}, fmt.Errorf("client_id required")
	}
	if len(req.SpendTx) == 0 {
		return CreateResp{}, fmt.Errorf("spend_tx required")
	}
	if req.InputAmount == 0 {
		return CreateResp{}, fmt.Errorf("input_amount required")
	}
	if req.SequenceNumber == 0 {
		return CreateResp{}, fmt.Errorf("sequence_number must be >= 1")
	}
	clientBSVPubHex, err := Libp2pMarshalPubHexToSecpCompressedHex(req.ClientID)
	if err != nil {
		return CreateResp{}, err
	}
	clientPub, err := ec.PublicKeyFromString(clientBSVPubHex)
	if err != nil {
		return CreateResp{}, fmt.Errorf("invalid client secp256k1 pubkey: %w", err)
	}
	serverActor, err := BuildActor("gateway", strings.TrimSpace(s.ServerPrivHex), s.IsMainnet)
	if err != nil {
		return CreateResp{}, err
	}

	spendTxHex := strings.ToLower(hex.EncodeToString(req.SpendTx))
	spendTx, err := tx.NewTransactionFromHex(spendTxHex)
	if err != nil {
		return CreateResp{}, fmt.Errorf("parse spend tx: %w", err)
	}
	clientSig := append([]byte(nil), req.ClientSig...)
	if len(clientSig) == 0 {
		return CreateResp{}, fmt.Errorf("client_signature required")
	}
	ok, err := ce.ServerVerifyClientSpendSig(spendTx, req.InputAmount, serverActor.PubKey, clientPub, &clientSig)
	if err != nil {
		return CreateResp{}, fmt.Errorf("verify client spend sig: %w", err)
	}
	if !ok {
		return CreateResp{}, fmt.Errorf("client spend signature invalid")
	}
	serverSig, err := ce.SpendTXServerSign(spendTx, req.InputAmount, serverActor.PrivKey, clientPub)
	if err != nil {
		return CreateResp{}, fmt.Errorf("server sign spend tx: %w", err)
	}
	mergedTx, err := ce.MergeDualPoolSigForSpendTx(spendTx.Hex(), serverSig, &clientSig)
	if err != nil {
		return CreateResp{}, fmt.Errorf("merge create signatures failed: %w", err)
	}

	spendTxID := spendTx.TxID().String()
	spendTxFee := CalcFeeWithInputAmount(spendTx, req.InputAmount)
	poolAmount := req.InputAmount
	if s.Params.MinimumPoolAmountSatoshi > 0 && poolAmount < s.Params.MinimumPoolAmountSatoshi {
		return CreateResp{}, fmt.Errorf("pool amount %d < minimum_pool_amount %d", poolAmount, s.Params.MinimumPoolAmountSatoshi)
	}
	if req.ServerAmount+spendTxFee > poolAmount {
		return CreateResp{}, fmt.Errorf("invalid initial amounts: server_amount %d + fee %d > pool_amount %d", req.ServerAmount, spendTxFee, poolAmount)
	}

	// 幂等：若 spend_txid 已存在，则直接返回（以数据库为准）。
		if old, found, loadErr := LoadSessionBySpendTxID(s.DB, spendTxID); loadErr == nil && found {
			return CreateResp{
				SpendTxID:     old.SpendTxID,
				ServerSig:     append([]byte(nil), *serverSig...),
				SpendTxFeeSat: old.SpendTxFeeSat,
				PoolAmountSat: old.PoolAmountSat,
			}, nil
		}

	row := GatewaySessionRow{
		SpendTxID:                 spendTxID,
		ClientID:                  strings.ToLower(strings.TrimSpace(req.ClientID)),
		ClientBSVCompressedPubHex: clientBSVPubHex,
		ServerBSVCompressedPubHex: strings.ToLower(serverActor.PubHex),
		InputAmountSat:            req.InputAmount,
		PoolAmountSat:             poolAmount,
		SpendTxFeeSat:             spendTxFee,
		Sequence:                  req.SequenceNumber,
		ServerAmountSat:           req.ServerAmount,
		ClientAmountSat:           poolAmount - req.ServerAmount - spendTxFee,
		BaseTxID:                  "",
		FinalTxID:                 "",
		BaseTxHex:                 "",
		CurrentTxHex:              mergedTx.Hex(),
		Status:                    "pending_base_tx",
	}
	if err := InsertSession(s.DB, row); err != nil {
		return CreateResp{}, err
	}
	obs.Business("bitcast-gateway", "fee_pool_create_ok", map[string]any{
		"spend_txid":   spendTxID,
		"pool_amount":  poolAmount,
		"spend_tx_fee": spendTxFee,
	})
	return CreateResp{
		SpendTxID:     spendTxID,
		ServerSig:     append([]byte(nil), *serverSig...),
		ErrorMessage:  "",
		SpendTxFeeSat: spendTxFee,
		PoolAmountSat: poolAmount,
	}, nil
}

func (s *GatewayService) BaseTx(req BaseTxReq) (BaseTxResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.DB == nil || s.Chain == nil {
		return BaseTxResp{}, fmt.Errorf("service not initialized")
	}
	if strings.TrimSpace(req.ClientID) == "" {
		return BaseTxResp{}, fmt.Errorf("client_id required")
	}
	row, found, err := LoadSessionBySpendTxID(s.DB, req.SpendTxID)
	if err != nil {
		return BaseTxResp{}, err
	}
	if !found {
		return BaseTxResp{Success: false, Status: "not_found", Error: "session not found by spend_txid"}, nil
	}
	if row.Status != "pending_base_tx" && row.Status != "active" {
		return BaseTxResp{Success: false, Status: row.Status, Error: "session status does not allow base_tx"}, nil
	}
	if row.BaseTxID != "" && row.Status == "active" {
		return BaseTxResp{Success: true, Status: "active", BaseTxID: row.BaseTxID}, nil
	}
	if err := s.rejectIfExpiredForPay(row); err != nil {
		return BaseTxResp{Success: false, Status: row.Status, Error: err.Error()}, nil
	}

	clientPub, err := ec.PublicKeyFromString(row.ClientBSVCompressedPubHex)
	if err != nil {
		return BaseTxResp{}, fmt.Errorf("invalid stored client pubkey: %w", err)
	}
	serverActor, err := BuildActor("gateway", strings.TrimSpace(s.ServerPrivHex), s.IsMainnet)
	if err != nil {
		return BaseTxResp{}, err
	}

	if len(req.BaseTx) == 0 {
		return BaseTxResp{Success: false, Status: row.Status, Error: "base tx required"}, nil
	}
	baseTxHex := strings.ToLower(hex.EncodeToString(req.BaseTx))
	baseTx, err := tx.NewTransactionFromHex(baseTxHex)
	if err != nil {
		return BaseTxResp{}, fmt.Errorf("parse base tx: %w", err)
	}
	multisigScript, err := libs.Lock([]*ec.PublicKey{serverActor.PubKey, clientPub}, 2)
	if err != nil {
		return BaseTxResp{}, fmt.Errorf("build multisig script: %w", err)
	}
	if len(baseTx.Outputs) == 0 {
		return BaseTxResp{Success: false, Status: row.Status, Error: "base tx has no outputs"}, nil
	}
	if baseTx.Outputs[0].LockingScript.String() != multisigScript.String() {
		return BaseTxResp{Success: false, Status: row.Status, Error: "base tx output[0] locking script mismatch"}, nil
	}
	baseTxID, err := s.Chain.Broadcast(baseTxHex)
	if err != nil {
		return BaseTxResp{Success: false, Status: row.Status, Error: "broadcast base tx failed: " + err.Error()}, nil
	}
	row.BaseTxID = baseTxID
	row.BaseTxHex = baseTxHex
	row.Status = "active"
	if err := UpdateSession(s.DB, row); err != nil {
		return BaseTxResp{}, err
	}
	obs.Business("bitcast-gateway", "fee_pool_base_tx_broadcasted", map[string]any{
		"spend_txid": row.SpendTxID,
		"base_txid":  baseTxID,
	})
	return BaseTxResp{Success: true, Status: "active", BaseTxID: baseTxID}, nil
}

func (s *GatewayService) PayConfirm(req PayConfirmReq) (PayConfirmResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.DB == nil {
		return PayConfirmResp{}, fmt.Errorf("db not initialized")
	}
	row, found, err := LoadSessionBySpendTxID(s.DB, req.SpendTxID)
	if err != nil {
		return PayConfirmResp{}, err
	}
	if !found {
		return PayConfirmResp{Success: false, Status: "not_found", Error: "session not found by spend_txid"}, nil
	}
	if row.Status != "active" {
		return PayConfirmResp{Success: false, Status: row.Status, Error: "session status does not allow pay"}, nil
	}
	if err := s.rejectIfExpiredForPay(row); err != nil {
		return PayConfirmResp{Success: false, Status: row.Status, Error: err.Error()}, nil
	}
	if req.SequenceNumber <= row.Sequence {
		return PayConfirmResp{Success: false, Status: row.Status, Error: fmt.Sprintf("sequence must increase: got=%d current=%d", req.SequenceNumber, row.Sequence)}, nil
	}
	if req.ServerAmount <= row.ServerAmountSat {
		return PayConfirmResp{Success: false, Status: row.Status, Error: fmt.Sprintf("server_amount must increase: got=%d current=%d", req.ServerAmount, row.ServerAmountSat)}, nil
	}
	if req.Fee != row.SpendTxFeeSat {
		return PayConfirmResp{Success: false, Status: row.Status, Error: fmt.Sprintf("fee mismatch: got=%d expect=%d", req.Fee, row.SpendTxFeeSat)}, nil
	}
	clientPub, err := ec.PublicKeyFromString(row.ClientBSVCompressedPubHex)
	if err != nil {
		return PayConfirmResp{}, fmt.Errorf("invalid stored client pubkey: %w", err)
	}
	serverActor, err := BuildActor("gateway", strings.TrimSpace(s.ServerPrivHex), s.IsMainnet)
	if err != nil {
		return PayConfirmResp{}, err
	}
	updatedTx, err := ce.LoadTx(
		row.CurrentTxHex,
		nil,
		req.SequenceNumber,
		req.ServerAmount,
		serverActor.PubKey,
		clientPub,
		row.PoolAmountSat,
	)
	if err != nil {
		return PayConfirmResp{Success: false, Status: row.Status, Error: "rebuild updated tx failed: " + err.Error()}, nil
	}
	clientSig := append([]byte(nil), req.ClientSig...)
	if len(clientSig) == 0 {
		return PayConfirmResp{Success: false, Status: row.Status, Error: "signature required"}, nil
	}
	ok, err := ce.ServerVerifyClientUpdateSig(updatedTx, serverActor.PubKey, clientPub, &clientSig)
	if err != nil {
		return PayConfirmResp{Success: false, Status: row.Status, Error: "verify client update sig failed: " + err.Error()}, nil
	}
	if !ok {
		return PayConfirmResp{Success: false, Status: row.Status, Error: "client update signature invalid"}, nil
	}
	serverSig, err := ce.SpendTXServerSign(updatedTx, row.PoolAmountSat, serverActor.PrivKey, clientPub)
	if err != nil {
		return PayConfirmResp{Success: false, Status: row.Status, Error: "server sign failed: " + err.Error()}, nil
	}
	mergedTx, err := ce.MergeDualPoolSigForSpendTx(updatedTx.Hex(), serverSig, &clientSig)
	if err != nil {
		return PayConfirmResp{Success: false, Status: row.Status, Error: "merge signatures failed: " + err.Error()}, nil
	}
	row.CurrentTxHex = mergedTx.Hex()
	row.Sequence = req.SequenceNumber
	row.ServerAmountSat = req.ServerAmount
	row.ClientAmountSat = row.PoolAmountSat - row.ServerAmountSat - row.SpendTxFeeSat
	if err := UpdateSession(s.DB, row); err != nil {
		return PayConfirmResp{}, err
	}
	chargeReason := strings.TrimSpace(req.ChargeReason)
	if chargeReason == "" {
		chargeReason = "unspecified"
	}
	if err := InsertChargeEvent(s.DB, row.ClientID, row.SpendTxID, row.Sequence, chargeReason, req.ChargeAmountSatoshi); err != nil {
		return PayConfirmResp{}, err
	}
	updatedTxID := mergedTx.TxID().String()
	obs.Business("bitcast-gateway", "fee_pool_pay_confirm_ok", map[string]any{
		"spend_txid":    row.SpendTxID,
		"updated_txid":  updatedTxID,
		"sequence":      row.Sequence,
		"server_amount": row.ServerAmountSat,
		"reason":        req.ChargeReason,
		"amount":        req.ChargeAmountSatoshi,
	})
	return PayConfirmResp{
		Success:      true,
		Status:       "active",
		UpdatedTxID:  updatedTxID,
		Sequence:     row.Sequence,
		ServerAmount: row.ServerAmountSat,
		ClientAmount: row.ClientAmountSat,
	}, nil
}

func (s *GatewayService) Close(req CloseReq) (CloseResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.DB == nil || s.Chain == nil {
		return CloseResp{}, fmt.Errorf("service not initialized")
	}
	row, found, err := LoadSessionBySpendTxID(s.DB, req.SpendTxID)
	if err != nil {
		return CloseResp{}, err
	}
	if !found {
		return CloseResp{Success: false, Status: "not_found", Error: "session not found by spend_txid"}, nil
	}
	if row.Status == "closed" && row.FinalTxID != "" {
		return CloseResp{Success: true, Status: "closed", Broadcasted: true, FinalSpendTxID: row.FinalTxID}, nil
	}
	if row.Status != "active" && row.Status != "pending_base_tx" {
		return CloseResp{Success: false, Status: row.Status, Error: "session status does not allow close"}, nil
	}
	if req.ServerAmount != row.ServerAmountSat {
		return CloseResp{Success: false, Status: row.Status, Error: fmt.Sprintf("server_amount mismatch: got=%d expect=%d", req.ServerAmount, row.ServerAmountSat)}, nil
	}
	if req.Fee != row.SpendTxFeeSat {
		return CloseResp{Success: false, Status: row.Status, Error: fmt.Sprintf("fee mismatch: got=%d expect=%d", req.Fee, row.SpendTxFeeSat)}, nil
	}
	clientPub, err := ec.PublicKeyFromString(row.ClientBSVCompressedPubHex)
	if err != nil {
		return CloseResp{}, fmt.Errorf("invalid stored client pubkey: %w", err)
	}
	serverActor, err := BuildActor("gateway", strings.TrimSpace(s.ServerPrivHex), s.IsMainnet)
	if err != nil {
		return CloseResp{}, err
	}
	finalLockTime := uint32(0xffffffff)
	finalSequence := uint32(0xffffffff)
	finalTx, err := ce.LoadTx(
		row.CurrentTxHex,
		&finalLockTime,
		finalSequence,
		req.ServerAmount,
		serverActor.PubKey,
		clientPub,
		row.PoolAmountSat,
	)
	if err != nil {
		return CloseResp{Success: false, Status: row.Status, Error: "rebuild final tx failed: " + err.Error()}, nil
	}
	clientSig := append([]byte(nil), req.ClientSig...)
	if len(clientSig) == 0 {
		return CloseResp{Success: false, Status: row.Status, Error: "signature required"}, nil
	}
	ok, err := ce.ServerVerifyClientSpendSig(finalTx, row.PoolAmountSat, serverActor.PubKey, clientPub, &clientSig)
	if err != nil {
		return CloseResp{Success: false, Status: row.Status, Error: "verify client final sig failed: " + err.Error()}, nil
	}
	if !ok {
		return CloseResp{Success: false, Status: row.Status, Error: "client final signature invalid"}, nil
	}
	serverSig, err := ce.SpendTXServerSign(finalTx, row.PoolAmountSat, serverActor.PrivKey, clientPub)
	if err != nil {
		return CloseResp{Success: false, Status: row.Status, Error: "server sign final failed: " + err.Error()}, nil
	}
	mergedTx, err := ce.MergeDualPoolSigForSpendTx(finalTx.Hex(), serverSig, &clientSig)
	if err != nil {
		return CloseResp{Success: false, Status: row.Status, Error: "merge final signatures failed: " + err.Error()}, nil
	}
	finalTxID, err := s.Chain.Broadcast(mergedTx.Hex())
	if err != nil {
		return CloseResp{Success: false, Status: row.Status, Error: "broadcast final tx failed: " + err.Error()}, nil
	}
	row.FinalTxID = finalTxID
	row.Status = "closed"
	if err := UpdateSession(s.DB, row); err != nil {
		return CloseResp{}, err
	}
	obs.Business("bitcast-gateway", "fee_pool_close_broadcasted", map[string]any{
		"spend_txid": row.SpendTxID,
		"final_txid": finalTxID,
	})
	return CloseResp{Success: true, Status: "closed", Broadcasted: true, FinalSpendTxID: finalTxID}, nil
}

func (s *GatewayService) State(req StateReq) (StateResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.DB == nil {
		return StateResp{}, fmt.Errorf("db not initialized")
	}
	clientID := strings.ToLower(strings.TrimSpace(req.ClientID))
	if clientID == "" {
		return StateResp{}, fmt.Errorf("client_id required")
	}
	var row GatewaySessionRow
	var found bool
	var err error
	if strings.TrimSpace(req.SpendTxID) != "" {
		row, found, err = LoadSessionBySpendTxID(s.DB, req.SpendTxID)
	} else {
		row, found, err = LoadLatestSessionByClientID(s.DB, clientID)
	}
	if err != nil {
		return StateResp{}, err
	}
	if !found {
		return StateResp{Status: "not_found"}, nil
	}
	currentTx, _ := hex.DecodeString(strings.TrimSpace(row.CurrentTxHex))
	return StateResp{
		Status:          row.Status,
		SpendTxID:       row.SpendTxID,
		BaseTxID:        row.BaseTxID,
		FinalTxID:       row.FinalTxID,
		CurrentTx:       currentTx,
		PoolAmountSat:   row.PoolAmountSat,
		SpendTxFeeSat:   row.SpendTxFeeSat,
		Sequence:        row.Sequence,
		ServerAmountSat: row.ServerAmountSat,
		ClientAmountSat: row.ClientAmountSat,
	}, nil
}

func (s *GatewayService) RunPassiveCloseLoop(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		if err := s.PassiveCloseExpiredOnce(); err != nil {
			obs.Error("bitcast-gateway", "fee_pool_passive_close_tick_failed", map[string]any{"error": err.Error()})
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (s *GatewayService) PassiveCloseExpiredOnce() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.DB == nil || s.Chain == nil {
		return fmt.Errorf("service not initialized")
	}
	tip, err := s.Chain.GetTipHeight()
	if err != nil {
		return fmt.Errorf("query tip height failed: %w", err)
	}
	rows, err := ListActiveSessions(s.DB)
	if err != nil {
		return err
	}
	for _, row := range rows {
		expireHeight, hasHeight, parseErr := sessionExpireHeight(row.CurrentTxHex)
		if parseErr != nil {
			obs.Error("bitcast-gateway", "fee_pool_passive_close_parse_failed", map[string]any{
				"spend_txid": row.SpendTxID,
				"error":      parseErr.Error(),
			})
			continue
		}
		if !hasHeight || tip < expireHeight {
			continue
		}
		finalTxID, bErr := s.Chain.Broadcast(row.CurrentTxHex)
		if bErr != nil {
			obs.Error("bitcast-gateway", "fee_pool_passive_close_broadcast_failed", map[string]any{
				"spend_txid":      row.SpendTxID,
				"tip_height":      tip,
				"expire_height":   expireHeight,
				"broadcast_error": bErr.Error(),
			})
			continue
		}
		row.FinalTxID = finalTxID
		row.Status = "closed"
		if uErr := UpdateSession(s.DB, row); uErr != nil {
			obs.Error("bitcast-gateway", "fee_pool_passive_close_update_failed", map[string]any{
				"spend_txid": row.SpendTxID,
				"final_txid": finalTxID,
				"error":      uErr.Error(),
			})
			continue
		}
		obs.Business("bitcast-gateway", "fee_pool_passive_close_broadcasted", map[string]any{
			"spend_txid":    row.SpendTxID,
			"final_txid":    finalTxID,
			"tip_height":    tip,
			"expire_height": expireHeight,
		})
	}
	return nil
}

func (s *GatewayService) rejectIfExpiredForPay(row GatewaySessionRow) error {
	if s.Chain == nil {
		return nil
	}
	expireHeight, hasHeight, err := sessionExpireHeight(row.CurrentTxHex)
	if err != nil {
		return fmt.Errorf("parse session spend tx failed: %w", err)
	}
	if !hasHeight {
		return nil
	}
	tip, err := s.Chain.GetTipHeight()
	if err != nil {
		return fmt.Errorf("query tip height failed: %w", err)
	}
	if tip >= expireHeight {
		return fmt.Errorf("fee pool expired at height=%d current_tip=%d", expireHeight, tip)
	}
	guardBlocks := s.Params.PayRejectBeforeExpiryBlocks
	if guardBlocks == 0 {
		guardBlocks = defaultPayGuardBlocks
	}
	if expireHeight > tip {
		remain := expireHeight - tip
		if remain <= guardBlocks {
			return fmt.Errorf("fee pool near expiry: expire_height=%d current_tip=%d guard_blocks=%d", expireHeight, tip, guardBlocks)
		}
	}
	return nil
}

func sessionExpireHeight(txHex string) (uint32, bool, error) {
	spendTxHex := strings.TrimSpace(txHex)
	if spendTxHex == "" {
		return 0, false, fmt.Errorf("empty current tx hex")
	}
	parsed, err := tx.NewTransactionFromHex(spendTxHex)
	if err != nil {
		return 0, false, err
	}
	lockTime := parsed.LockTime
	if lockTime == 0 || lockTime == 0xffffffff {
		return 0, false, nil
	}
	if lockTime >= feePoolLockTimeThreshold {
		return 0, false, nil
	}
	return lockTime, true, nil
}
