package broadcast

import oldproto "github.com/golang/protobuf/proto"

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
