package poolcore

import "fmt"

// ReceiptFinalizeSpec 描述一次已扣费业务结果如何补齐公共回执骨架。
// 约束：
// - 业务层自己决定 payload 长什么样、result_code 取什么值；
// - poolcore 只负责把 payConfirm 结果和标准 service receipt 拼进去；
// - 这样共享的是“骨架”，不是业务结果对象本身。
type ReceiptFinalizeSpec[T any] struct {
	ServiceType       string
	ApplyPayOut       func(T, PayConfirmResp) T
	MarshalPayload    func(T) ([]byte, error)
	ResultCode        func(T) string
	SetServiceReceipt func(T, []byte) T
}

func FinalizeServiceResult[T any](serverPrivHex string, isMainnet bool, clientID string, payOut PayConfirmResp, resp T, spec ReceiptFinalizeSpec[T]) (T, error) {
	if spec.ApplyPayOut == nil {
		return resp, fmt.Errorf("apply payout handler required")
	}
	if spec.MarshalPayload == nil {
		return resp, fmt.Errorf("marshal payload handler required")
	}
	if spec.SetServiceReceipt == nil {
		return resp, fmt.Errorf("set service receipt handler required")
	}
	resp = spec.ApplyPayOut(resp, payOut)
	payload, err := spec.MarshalPayload(resp)
	if err != nil {
		return resp, err
	}
	resultCode := ""
	if spec.ResultCode != nil {
		resultCode = spec.ResultCode(resp)
	}
	receipt, err := BuildSignedServiceReceipt(serverPrivHex, isMainnet, clientID, payOut, spec.ServiceType, resultCode, payload)
	if err != nil {
		return resp, err
	}
	return spec.SetServiceReceipt(resp, receipt), nil
}
