package domainmodule

import (
	"bytes"
	"testing"

	txsdk "github.com/bsv-blockchain/go-sdk/transaction"
)

func TestSignedEnvelopeOpReturnRoundTrip(t *testing.T) {
	t.Parallel()

	payload := []byte(`[[["v","quote-id","alice"],"sig"]]`)
	s, err := BuildSignedEnvelopeOpReturnScript(payload)
	if err != nil {
		t.Fatalf("BuildSignedEnvelopeOpReturnScript() error = %v", err)
	}
	got, err := ExtractSignedEnvelopeOpReturnPayload(s)
	if err != nil {
		t.Fatalf("ExtractSignedEnvelopeOpReturnPayload() error = %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("payload mismatch: got=%q want=%q", got, payload)
	}
}

func TestExtractSignedEnvelopeOpReturnPayload_FromSDKOutput(t *testing.T) {
	t.Parallel()

	payload := []byte(`[[["v","quote-id","alice"],"sig"]]`)
	tx := txsdk.NewTransaction()
	if err := tx.AddOpReturnOutput(payload); err != nil {
		t.Fatalf("AddOpReturnOutput() error = %v", err)
	}
	got, err := ExtractSignedEnvelopeOpReturnPayload(tx.Outputs[0].LockingScript)
	if err != nil {
		t.Fatalf("ExtractSignedEnvelopeOpReturnPayload() error = %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("payload mismatch: got=%q want=%q", got, payload)
	}
}
