package ncall

import oldproto "github.com/golang/protobuf/proto"

func (m *CallReq) Reset()         { *m = CallReq{} }
func (m *CallReq) String() string { return oldproto.CompactTextString(m) }
func (*CallReq) ProtoMessage()    {}

func (m *PaymentOption) Reset()         { *m = PaymentOption{} }
func (m *PaymentOption) String() string { return oldproto.CompactTextString(m) }
func (*PaymentOption) ProtoMessage()    {}

func (m *CallResp) Reset()         { *m = CallResp{} }
func (m *CallResp) String() string { return oldproto.CompactTextString(m) }
func (*CallResp) ProtoMessage()    {}

func (m *ResolveReq) Reset()         { *m = ResolveReq{} }
func (m *ResolveReq) String() string { return oldproto.CompactTextString(m) }
func (*ResolveReq) ProtoMessage()    {}

func (m *ResolveResp) Reset()         { *m = ResolveResp{} }
func (m *ResolveResp) String() string { return oldproto.CompactTextString(m) }
func (*ResolveResp) ProtoMessage()    {}

func (m *FeePool2of2Payment) Reset()         { *m = FeePool2of2Payment{} }
func (m *FeePool2of2Payment) String() string { return oldproto.CompactTextString(m) }
func (*FeePool2of2Payment) ProtoMessage()    {}

func (m *FeePool2of2Receipt) Reset()         { *m = FeePool2of2Receipt{} }
func (m *FeePool2of2Receipt) String() string { return oldproto.CompactTextString(m) }
func (*FeePool2of2Receipt) ProtoMessage()    {}

func (m *CapabilityItem) Reset()         { *m = CapabilityItem{} }
func (m *CapabilityItem) String() string { return oldproto.CompactTextString(m) }
func (*CapabilityItem) ProtoMessage()    {}

func (m *CapabilitiesShowBody) Reset()         { *m = CapabilitiesShowBody{} }
func (m *CapabilitiesShowBody) String() string { return oldproto.CompactTextString(m) }
func (*CapabilitiesShowBody) ProtoMessage()    {}
