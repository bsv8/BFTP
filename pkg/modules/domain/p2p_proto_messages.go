package domainsvc

import oldproto "github.com/golang/protobuf/proto"

func (m *NameRouteReq) Reset()         { *m = NameRouteReq{} }
func (m *NameRouteReq) String() string { return oldproto.CompactTextString(m) }
func (*NameRouteReq) ProtoMessage()    {}

func (m *NameTargetRouteReq) Reset()         { *m = NameTargetRouteReq{} }
func (m *NameTargetRouteReq) String() string { return oldproto.CompactTextString(m) }
func (*NameTargetRouteReq) ProtoMessage()    {}

func (m *DomainPricingBody) Reset()         { *m = DomainPricingBody{} }
func (m *DomainPricingBody) String() string { return oldproto.CompactTextString(m) }
func (*DomainPricingBody) ProtoMessage()    {}

func (m *ListOwnedReq) Reset()         { *m = ListOwnedReq{} }
func (m *ListOwnedReq) String() string { return oldproto.CompactTextString(m) }
func (*ListOwnedReq) ProtoMessage()    {}

func (m *OwnedNameItem) Reset()         { *m = OwnedNameItem{} }
func (m *OwnedNameItem) String() string { return oldproto.CompactTextString(m) }
func (*OwnedNameItem) ProtoMessage()    {}

func (m *ListOwnedResp) Reset()         { *m = ListOwnedResp{} }
func (m *ListOwnedResp) String() string { return oldproto.CompactTextString(m) }
func (*ListOwnedResp) ProtoMessage()    {}

func (m *ResolveNamePaidReq) Reset()         { *m = ResolveNamePaidReq{} }
func (m *ResolveNamePaidReq) String() string { return oldproto.CompactTextString(m) }
func (*ResolveNamePaidReq) ProtoMessage()    {}

func (m *ResolveNamePaidResp) Reset()         { *m = ResolveNamePaidResp{} }
func (m *ResolveNamePaidResp) String() string { return oldproto.CompactTextString(m) }
func (*ResolveNamePaidResp) ProtoMessage()    {}

func (m *QueryNamePaidReq) Reset()         { *m = QueryNamePaidReq{} }
func (m *QueryNamePaidReq) String() string { return oldproto.CompactTextString(m) }
func (*QueryNamePaidReq) ProtoMessage()    {}

func (m *QueryNamePaidResp) Reset()         { *m = QueryNamePaidResp{} }
func (m *QueryNamePaidResp) String() string { return oldproto.CompactTextString(m) }
func (*QueryNamePaidResp) ProtoMessage()    {}

func (m *RegisterLockPaidReq) Reset()         { *m = RegisterLockPaidReq{} }
func (m *RegisterLockPaidReq) String() string { return oldproto.CompactTextString(m) }
func (*RegisterLockPaidReq) ProtoMessage()    {}

func (m *RegisterLockPaidResp) Reset()         { *m = RegisterLockPaidResp{} }
func (m *RegisterLockPaidResp) String() string { return oldproto.CompactTextString(m) }
func (*RegisterLockPaidResp) ProtoMessage()    {}

func (m *RegisterSubmitReq) Reset()         { *m = RegisterSubmitReq{} }
func (m *RegisterSubmitReq) String() string { return oldproto.CompactTextString(m) }
func (*RegisterSubmitReq) ProtoMessage()    {}

func (m *RegisterSubmitResp) Reset()         { *m = RegisterSubmitResp{} }
func (m *RegisterSubmitResp) String() string { return oldproto.CompactTextString(m) }
func (*RegisterSubmitResp) ProtoMessage()    {}

func (m *SetTargetPaidReq) Reset()         { *m = SetTargetPaidReq{} }
func (m *SetTargetPaidReq) String() string { return oldproto.CompactTextString(m) }
func (*SetTargetPaidReq) ProtoMessage()    {}

func (m *SetTargetPaidResp) Reset()         { *m = SetTargetPaidResp{} }
func (m *SetTargetPaidResp) String() string { return oldproto.CompactTextString(m) }
func (*SetTargetPaidResp) ProtoMessage()    {}
