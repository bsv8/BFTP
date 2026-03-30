package broadcast

import oldproto "github.com/golang/protobuf/proto"

func (m *DemandPublishReq) Reset()         { *m = DemandPublishReq{} }
func (m *DemandPublishReq) String() string { return oldproto.CompactTextString(m) }
func (*DemandPublishReq) ProtoMessage()    {}

func (m *DemandPublishBatchReq) Reset()         { *m = DemandPublishBatchReq{} }
func (m *DemandPublishBatchReq) String() string { return oldproto.CompactTextString(m) }
func (*DemandPublishBatchReq) ProtoMessage()    {}

func (m *LiveDemandPublishReq) Reset()         { *m = LiveDemandPublishReq{} }
func (m *LiveDemandPublishReq) String() string { return oldproto.CompactTextString(m) }
func (*LiveDemandPublishReq) ProtoMessage()    {}

func (m *NodeReachabilityAnnounceReq) Reset()         { *m = NodeReachabilityAnnounceReq{} }
func (m *NodeReachabilityAnnounceReq) String() string { return oldproto.CompactTextString(m) }
func (*NodeReachabilityAnnounceReq) ProtoMessage()    {}

func (m *NodeReachabilityQueryReq) Reset()         { *m = NodeReachabilityQueryReq{} }
func (m *NodeReachabilityQueryReq) String() string { return oldproto.CompactTextString(m) }
func (*NodeReachabilityQueryReq) ProtoMessage()    {}

func (m *DemandPublishPaidResp) Reset()         { *m = DemandPublishPaidResp{} }
func (m *DemandPublishPaidResp) String() string { return oldproto.CompactTextString(m) }
func (*DemandPublishPaidResp) ProtoMessage()    {}

func (m *DemandPublishBatchPaidItem) Reset()         { *m = DemandPublishBatchPaidItem{} }
func (m *DemandPublishBatchPaidItem) String() string { return oldproto.CompactTextString(m) }
func (*DemandPublishBatchPaidItem) ProtoMessage()    {}

func (m *DemandPublishBatchPaidResult) Reset()         { *m = DemandPublishBatchPaidResult{} }
func (m *DemandPublishBatchPaidResult) String() string { return oldproto.CompactTextString(m) }
func (*DemandPublishBatchPaidResult) ProtoMessage()    {}

func (m *DemandPublishBatchPaidResp) Reset()         { *m = DemandPublishBatchPaidResp{} }
func (m *DemandPublishBatchPaidResp) String() string { return oldproto.CompactTextString(m) }
func (*DemandPublishBatchPaidResp) ProtoMessage()    {}

func (m *LiveDemandPublishPaidResp) Reset()         { *m = LiveDemandPublishPaidResp{} }
func (m *LiveDemandPublishPaidResp) String() string { return oldproto.CompactTextString(m) }
func (*LiveDemandPublishPaidResp) ProtoMessage()    {}

func (m *NodeReachabilityAnnouncePaidResp) Reset()         { *m = NodeReachabilityAnnouncePaidResp{} }
func (m *NodeReachabilityAnnouncePaidResp) String() string { return oldproto.CompactTextString(m) }
func (*NodeReachabilityAnnouncePaidResp) ProtoMessage()    {}

func (m *NodeReachabilityQueryPaidResp) Reset()         { *m = NodeReachabilityQueryPaidResp{} }
func (m *NodeReachabilityQueryPaidResp) String() string { return oldproto.CompactTextString(m) }
func (*NodeReachabilityQueryPaidResp) ProtoMessage()    {}
