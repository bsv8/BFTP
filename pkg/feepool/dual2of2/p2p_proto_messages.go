package dual2of2

import oldproto "github.com/golang/protobuf/proto"

func (m *InfoReq) Reset()         { *m = InfoReq{} }
func (m *InfoReq) String() string { return oldproto.CompactTextString(m) }
func (*InfoReq) ProtoMessage()    {}

func (m *InfoResp) Reset()         { *m = InfoResp{} }
func (m *InfoResp) String() string { return oldproto.CompactTextString(m) }
func (*InfoResp) ProtoMessage()    {}

func (m *CreateReq) Reset()         { *m = CreateReq{} }
func (m *CreateReq) String() string { return oldproto.CompactTextString(m) }
func (*CreateReq) ProtoMessage()    {}

func (m *CreateResp) Reset()         { *m = CreateResp{} }
func (m *CreateResp) String() string { return oldproto.CompactTextString(m) }
func (*CreateResp) ProtoMessage()    {}

func (m *BaseTxReq) Reset()         { *m = BaseTxReq{} }
func (m *BaseTxReq) String() string { return oldproto.CompactTextString(m) }
func (*BaseTxReq) ProtoMessage()    {}

func (m *BaseTxResp) Reset()         { *m = BaseTxResp{} }
func (m *BaseTxResp) String() string { return oldproto.CompactTextString(m) }
func (*BaseTxResp) ProtoMessage()    {}

func (m *PayConfirmReq) Reset()         { *m = PayConfirmReq{} }
func (m *PayConfirmReq) String() string { return oldproto.CompactTextString(m) }
func (*PayConfirmReq) ProtoMessage()    {}

func (m *PayConfirmResp) Reset()         { *m = PayConfirmResp{} }
func (m *PayConfirmResp) String() string { return oldproto.CompactTextString(m) }
func (*PayConfirmResp) ProtoMessage()    {}

func (m *CloseReq) Reset()         { *m = CloseReq{} }
func (m *CloseReq) String() string { return oldproto.CompactTextString(m) }
func (*CloseReq) ProtoMessage()    {}

func (m *CloseResp) Reset()         { *m = CloseResp{} }
func (m *CloseResp) String() string { return oldproto.CompactTextString(m) }
func (*CloseResp) ProtoMessage()    {}

func (m *StateReq) Reset()         { *m = StateReq{} }
func (m *StateReq) String() string { return oldproto.CompactTextString(m) }
func (*StateReq) ProtoMessage()    {}

func (m *StateResp) Reset()         { *m = StateResp{} }
func (m *StateResp) String() string { return oldproto.CompactTextString(m) }
func (*StateResp) ProtoMessage()    {}

func (m *DemandPublishPaidReq) Reset()         { *m = DemandPublishPaidReq{} }
func (m *DemandPublishPaidReq) String() string { return oldproto.CompactTextString(m) }
func (*DemandPublishPaidReq) ProtoMessage()    {}

func (m *DemandPublishPaidResp) Reset()         { *m = DemandPublishPaidResp{} }
func (m *DemandPublishPaidResp) String() string { return oldproto.CompactTextString(m) }
func (*DemandPublishPaidResp) ProtoMessage()    {}

func (m *DemandPublishBatchPaidItem) Reset()         { *m = DemandPublishBatchPaidItem{} }
func (m *DemandPublishBatchPaidItem) String() string { return oldproto.CompactTextString(m) }
func (*DemandPublishBatchPaidItem) ProtoMessage()    {}

func (m *DemandPublishBatchPaidReq) Reset()         { *m = DemandPublishBatchPaidReq{} }
func (m *DemandPublishBatchPaidReq) String() string { return oldproto.CompactTextString(m) }
func (*DemandPublishBatchPaidReq) ProtoMessage()    {}

func (m *DemandPublishBatchPaidResult) Reset()         { *m = DemandPublishBatchPaidResult{} }
func (m *DemandPublishBatchPaidResult) String() string { return oldproto.CompactTextString(m) }
func (*DemandPublishBatchPaidResult) ProtoMessage()    {}

func (m *DemandPublishBatchPaidResp) Reset()         { *m = DemandPublishBatchPaidResp{} }
func (m *DemandPublishBatchPaidResp) String() string { return oldproto.CompactTextString(m) }
func (*DemandPublishBatchPaidResp) ProtoMessage()    {}

func (m *LiveDemandPublishPaidReq) Reset()         { *m = LiveDemandPublishPaidReq{} }
func (m *LiveDemandPublishPaidReq) String() string { return oldproto.CompactTextString(m) }
func (*LiveDemandPublishPaidReq) ProtoMessage()    {}

func (m *LiveDemandPublishPaidResp) Reset()         { *m = LiveDemandPublishPaidResp{} }
func (m *LiveDemandPublishPaidResp) String() string { return oldproto.CompactTextString(m) }
func (*LiveDemandPublishPaidResp) ProtoMessage()    {}

func (m *NodeReachabilityAnnouncePaidReq) Reset()         { *m = NodeReachabilityAnnouncePaidReq{} }
func (m *NodeReachabilityAnnouncePaidReq) String() string { return oldproto.CompactTextString(m) }
func (*NodeReachabilityAnnouncePaidReq) ProtoMessage()    {}

func (m *NodeReachabilityAnnouncePaidResp) Reset()         { *m = NodeReachabilityAnnouncePaidResp{} }
func (m *NodeReachabilityAnnouncePaidResp) String() string { return oldproto.CompactTextString(m) }
func (*NodeReachabilityAnnouncePaidResp) ProtoMessage()    {}

func (m *NodeReachabilityQueryPaidReq) Reset()         { *m = NodeReachabilityQueryPaidReq{} }
func (m *NodeReachabilityQueryPaidReq) String() string { return oldproto.CompactTextString(m) }
func (*NodeReachabilityQueryPaidReq) ProtoMessage()    {}

func (m *NodeReachabilityQueryPaidResp) Reset()         { *m = NodeReachabilityQueryPaidResp{} }
func (m *NodeReachabilityQueryPaidResp) String() string { return oldproto.CompactTextString(m) }
func (*NodeReachabilityQueryPaidResp) ProtoMessage()    {}
