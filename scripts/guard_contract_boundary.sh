#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

TYPE_LIST="CallReq|PaymentOption|CallResp|ResolveReq|ResolveResp|FeePool2of2Payment|FeePool2of2Receipt|ChainTxV1Payment|ChainTxV1Receipt|CapabilityItem|CapabilitiesShowBody|NameRouteReq|NameTargetRouteReq|DomainPricingBody|ListOwnedReq|OwnedNameItem|ListOwnedResp|ResolveNamePaidReq|ResolveNamePaidResp|QueryNamePaidReq|QueryNamePaidResp|RegisterLockPaidReq|RegisterLockPaidResp|RegisterSubmitReq|RegisterSubmitResp|SetTargetPaidReq|SetTargetPaidResp|DemandPublishReq|ListenCycleReq|DemandPublishBatchReq|LiveDemandPublishReq|NodeReachabilityAnnounceReq|NodeReachabilityQueryReq|DemandPublishPaidResp|ListenCyclePaidResp|DemandPublishBatchPaidItem|DemandPublishBatchPaidResult|DemandPublishBatchPaidResp|LiveDemandPublishPaidResp|NodeReachabilityAnnouncePaidResp|NodeReachabilityQueryPaidResp|ServiceOffer|ServiceQuote|ChargeIntent|ClientCommit|AcceptedCharge|ProofState|ServiceReceipt"

TARGET_DIRS=(
  "pkg/infra/ncall"
  "pkg/modules/domain"
  "pkg/modules/broadcast"
  "pkg/infra/payflow"
)

echo "[guard] checking contract boundary..."
if rg -n --glob '!**/*_test.go' "type (${TYPE_LIST}) struct" "${TARGET_DIRS[@]}"; then
  echo "[guard] boundary violated: contract struct must live in BFTP-contract" >&2
  exit 1
fi

if rg -n --glob '!**/*_test.go' 'Proto[A-Za-z0-9_]+\s+protocol\.ID\s*=\s*"/bsv-transfer/' pkg/infra/ncall pkg/modules/domain; then
  echo "[guard] boundary violated: protocol IDs must live in BFTP-contract/pkg/v1/protoid" >&2
  exit 1
fi

if rg -n --glob '!**/*_test.go' 'PaymentScheme[A-Za-z0-9_]+\s*=\s*"(pool_2of2_v1|chain_tx_v1)"' pkg/infra/ncall; then
  echo "[guard] boundary violated: payment schemes must live in BFTP-contract/pkg/v1/protoid" >&2
  exit 1
fi

if ! rg -q 'type ServiceOffer = contractpayflow.ServiceOffer' pkg/infra/payflow/types.go; then
  echo "[guard] payflow bridge missing: pkg/infra/payflow/types.go" >&2
  exit 1
fi

echo "[guard] ok"
