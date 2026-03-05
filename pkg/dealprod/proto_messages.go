package dealprod

import oldproto "github.com/golang/protobuf/proto"

func (m *PublishDemandReq) Reset()         { *m = PublishDemandReq{} }
func (m *PublishDemandReq) String() string { return oldproto.CompactTextString(m) }
func (*PublishDemandReq) ProtoMessage()    {}

func (m *PublishDemandResp) Reset()         { *m = PublishDemandResp{} }
func (m *PublishDemandResp) String() string { return oldproto.CompactTextString(m) }
func (*PublishDemandResp) ProtoMessage()    {}

func (m *QuoteSubmitReq) Reset()         { *m = QuoteSubmitReq{} }
func (m *QuoteSubmitReq) String() string { return oldproto.CompactTextString(m) }
func (*QuoteSubmitReq) ProtoMessage()    {}

func (m *QuoteSubmitResp) Reset()         { *m = QuoteSubmitResp{} }
func (m *QuoteSubmitResp) String() string { return oldproto.CompactTextString(m) }
func (*QuoteSubmitResp) ProtoMessage()    {}

func (m *QuoteListReq) Reset()         { *m = QuoteListReq{} }
func (m *QuoteListReq) String() string { return oldproto.CompactTextString(m) }
func (*QuoteListReq) ProtoMessage()    {}

func (m *QuoteItem) Reset()         { *m = QuoteItem{} }
func (m *QuoteItem) String() string { return oldproto.CompactTextString(m) }
func (*QuoteItem) ProtoMessage()    {}

func (m *QuoteListResp) Reset()         { *m = QuoteListResp{} }
func (m *QuoteListResp) String() string { return oldproto.CompactTextString(m) }
func (*QuoteListResp) ProtoMessage()    {}

func (m *DemandAnnounceReq) Reset()         { *m = DemandAnnounceReq{} }
func (m *DemandAnnounceReq) String() string { return oldproto.CompactTextString(m) }
func (*DemandAnnounceReq) ProtoMessage()    {}

func (m *DemandAnnounceResp) Reset()         { *m = DemandAnnounceResp{} }
func (m *DemandAnnounceResp) String() string { return oldproto.CompactTextString(m) }
func (*DemandAnnounceResp) ProtoMessage()    {}

func (m *DealAcceptReq) Reset()         { *m = DealAcceptReq{} }
func (m *DealAcceptReq) String() string { return oldproto.CompactTextString(m) }
func (*DealAcceptReq) ProtoMessage()    {}

func (m *DealAcceptResp) Reset()         { *m = DealAcceptResp{} }
func (m *DealAcceptResp) String() string { return oldproto.CompactTextString(m) }
func (*DealAcceptResp) ProtoMessage()    {}

func (m *SessionOpenReq) Reset()         { *m = SessionOpenReq{} }
func (m *SessionOpenReq) String() string { return oldproto.CompactTextString(m) }
func (*SessionOpenReq) ProtoMessage()    {}

func (m *SessionOpenResp) Reset()         { *m = SessionOpenResp{} }
func (m *SessionOpenResp) String() string { return oldproto.CompactTextString(m) }
func (*SessionOpenResp) ProtoMessage()    {}

func (m *ReleaseChunkReq) Reset()         { *m = ReleaseChunkReq{} }
func (m *ReleaseChunkReq) String() string { return oldproto.CompactTextString(m) }
func (*ReleaseChunkReq) ProtoMessage()    {}

func (m *ReleaseChunkResp) Reset()         { *m = ReleaseChunkResp{} }
func (m *ReleaseChunkResp) String() string { return oldproto.CompactTextString(m) }
func (*ReleaseChunkResp) ProtoMessage()    {}

func (m *VerifyReleaseReq) Reset()         { *m = VerifyReleaseReq{} }
func (m *VerifyReleaseReq) String() string { return oldproto.CompactTextString(m) }
func (*VerifyReleaseReq) ProtoMessage()    {}

func (m *VerifyReleaseResp) Reset()         { *m = VerifyReleaseResp{} }
func (m *VerifyReleaseResp) String() string { return oldproto.CompactTextString(m) }
func (*VerifyReleaseResp) ProtoMessage()    {}

func (m *SessionCloseReq) Reset()         { *m = SessionCloseReq{} }
func (m *SessionCloseReq) String() string { return oldproto.CompactTextString(m) }
func (*SessionCloseReq) ProtoMessage()    {}

func (m *SessionCloseResp) Reset()         { *m = SessionCloseResp{} }
func (m *SessionCloseResp) String() string { return oldproto.CompactTextString(m) }
func (*SessionCloseResp) ProtoMessage()    {}

func (m *DisputeReq) Reset()         { *m = DisputeReq{} }
func (m *DisputeReq) String() string { return oldproto.CompactTextString(m) }
func (*DisputeReq) ProtoMessage()    {}

func (m *DisputeResp) Reset()         { *m = DisputeResp{} }
func (m *DisputeResp) String() string { return oldproto.CompactTextString(m) }
func (*DisputeResp) ProtoMessage()    {}
