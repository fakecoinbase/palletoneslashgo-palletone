// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proposal.proto

package peer

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// This structure is necessary to sign the proposal which contains the header
// and the payload. Without this structure, we would have to concatenate the
// header and the payload to verify the signature, which could be expensive
// with large payload
//
// When an endorser receives a SignedProposal message, it should verify the
// signature over the proposal bytes. This verification requires the following
// steps:
// 1. Verification of the validity of the certificate that was used to produce
//    the signature.  The certificate will be available once proposalBytes has
//    been unmarshalled to a Proposal message, and Proposal.header has been
//    unmarshalled to a Header message. While this unmarshalling-before-verifying
//    might not be ideal, it is unavoidable because i) the signature needs to also
//    protect the signing certificate; ii) it is desirable that Header is created
//    once by the client and never changed (for the sake of accountability and
//    non-repudiation). Note also that it is actually impossible to conclusively
//    verify the validity of the certificate included in a Proposal, because the
//    proposal needs to first be endorsed and ordered with respect to certificate
//    expiration transactions. Still, it is useful to pre-filter expired
//    certificates at this stage.
// 2. Verification that the certificate is trusted (signed by a trusted CA) and
//    that it is allowed to transact with us (with respect to some ACLs);
// 3. Verification that the signature on proposalBytes is valid;
// 4. Detect replay attacks;
type PtnSignedProposal struct {
	// The bytes of Proposal
	ProposalBytes []byte `protobuf:"bytes,1,opt,name=proposal_bytes,json=proposalBytes,proto3" json:"proposal_bytes,omitempty"`
	// Signaure over proposalBytes; this signature is to be verified against
	// the creator identity contained in the header of the Proposal message
	// marshaled as proposalBytes
	Signature            []byte   `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PtnSignedProposal) Reset()         { *m = PtnSignedProposal{} }
func (m *PtnSignedProposal) String() string { return proto.CompactTextString(m) }
func (*PtnSignedProposal) ProtoMessage()    {}
func (*PtnSignedProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_c3ac5ce23bf32d05, []int{0}
}

func (m *PtnSignedProposal) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PtnSignedProposal.Unmarshal(m, b)
}
func (m *PtnSignedProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PtnSignedProposal.Marshal(b, m, deterministic)
}
func (m *PtnSignedProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PtnSignedProposal.Merge(m, src)
}
func (m *PtnSignedProposal) XXX_Size() int {
	return xxx_messageInfo_PtnSignedProposal.Size(m)
}
func (m *PtnSignedProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_PtnSignedProposal.DiscardUnknown(m)
}

var xxx_messageInfo_PtnSignedProposal proto.InternalMessageInfo

func (m *PtnSignedProposal) GetProposalBytes() []byte {
	if m != nil {
		return m.ProposalBytes
	}
	return nil
}

func (m *PtnSignedProposal) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

// A Proposal is sent to an endorser for endorsement.  The proposal contains:
// 1. A header which should be unmarshaled to a Header message.  Note that
//    Header is both the header of a Proposal and of a Transaction, in that i)
//    both headers should be unmarshaled to this message; and ii) it is used to
//    compute cryptographic hashes and signatures.  The header has fields common
//    to all proposals/transactions.  In addition it has a type field for
//    additional customization. An example of this is the ChaincodeHeaderExtension
//    message used to extend the Header for type CHAINCODE.
// 2. A payload whose type depends on the header's type field.
// 3. An extension whose type depends on the header's type field.
//
// Let us see an example. For type CHAINCODE (see the Header message),
// we have the following:
// 1. The header is a Header message whose extensions field is a
//    ChaincodeHeaderExtension message.
// 2. The payload is a ChaincodeProposalPayload message.
// 3. The extension is a ChaincodeAction that might be used to ask the
//    endorsers to endorse a specific ChaincodeAction, thus emulating the
//    submitting peer model.
type PtnProposal struct {
	// The header of the proposal. It is the bytes of the Header
	Header []byte `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	// The payload of the proposal as defined by the type in the proposal
	// header.
	Payload []byte `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
	// Optional extensions to the proposal. Its content depends on the Header's
	// type field.  For the type CHAINCODE, it might be the bytes of a
	// ChaincodeAction message.
	Extension            []byte   `protobuf:"bytes,3,opt,name=extension,proto3" json:"extension,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PtnProposal) Reset()         { *m = PtnProposal{} }
func (m *PtnProposal) String() string { return proto.CompactTextString(m) }
func (*PtnProposal) ProtoMessage()    {}
func (*PtnProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_c3ac5ce23bf32d05, []int{1}
}

func (m *PtnProposal) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PtnProposal.Unmarshal(m, b)
}
func (m *PtnProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PtnProposal.Marshal(b, m, deterministic)
}
func (m *PtnProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PtnProposal.Merge(m, src)
}
func (m *PtnProposal) XXX_Size() int {
	return xxx_messageInfo_PtnProposal.Size(m)
}
func (m *PtnProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_PtnProposal.DiscardUnknown(m)
}

var xxx_messageInfo_PtnProposal proto.InternalMessageInfo

func (m *PtnProposal) GetHeader() []byte {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *PtnProposal) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *PtnProposal) GetExtension() []byte {
	if m != nil {
		return m.Extension
	}
	return nil
}

// ChaincodeHeaderExtension is the Header's extentions message to be used when
// the Header's type is CHAINCODE.  This extensions is used to specify which
// chaincode to invoke and what should appear on the ledger.
type PtnChaincodeHeaderExtension struct {
	// The PayloadVisibility field controls to what extent the Proposal's payload
	// (recall that for the type CHAINCODE, it is ChaincodeProposalPayload
	// message) field will be visible in the final transaction and in the ledger.
	// Ideally, it would be configurable, supporting at least 3 main visibility
	// modes:
	// 1. all bytes of the payload are visible;
	// 2. only a hash of the payload is visible;
	// 3. nothing is visible.
	// Notice that the visibility function may be potentially part of the ESCC.
	// In that case it overrides PayloadVisibility field.  Finally notice that
	// this field impacts the content of ProposalResponsePayload.proposalHash.
	PayloadVisibility []byte `protobuf:"bytes,1,opt,name=payload_visibility,json=payloadVisibility,proto3" json:"payload_visibility,omitempty"`
	// The ID of the chaincode to target.
	ChaincodeId          *PtnChaincodeID `protobuf:"bytes,2,opt,name=chaincode_id,json=chaincodeId,proto3" json:"chaincode_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *PtnChaincodeHeaderExtension) Reset()         { *m = PtnChaincodeHeaderExtension{} }
func (m *PtnChaincodeHeaderExtension) String() string { return proto.CompactTextString(m) }
func (*PtnChaincodeHeaderExtension) ProtoMessage()    {}
func (*PtnChaincodeHeaderExtension) Descriptor() ([]byte, []int) {
	return fileDescriptor_c3ac5ce23bf32d05, []int{2}
}

func (m *PtnChaincodeHeaderExtension) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PtnChaincodeHeaderExtension.Unmarshal(m, b)
}
func (m *PtnChaincodeHeaderExtension) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PtnChaincodeHeaderExtension.Marshal(b, m, deterministic)
}
func (m *PtnChaincodeHeaderExtension) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PtnChaincodeHeaderExtension.Merge(m, src)
}
func (m *PtnChaincodeHeaderExtension) XXX_Size() int {
	return xxx_messageInfo_PtnChaincodeHeaderExtension.Size(m)
}
func (m *PtnChaincodeHeaderExtension) XXX_DiscardUnknown() {
	xxx_messageInfo_PtnChaincodeHeaderExtension.DiscardUnknown(m)
}

var xxx_messageInfo_PtnChaincodeHeaderExtension proto.InternalMessageInfo

func (m *PtnChaincodeHeaderExtension) GetPayloadVisibility() []byte {
	if m != nil {
		return m.PayloadVisibility
	}
	return nil
}

func (m *PtnChaincodeHeaderExtension) GetChaincodeId() *PtnChaincodeID {
	if m != nil {
		return m.ChaincodeId
	}
	return nil
}

// ChaincodeProposalPayload is the Proposal's payload message to be used when
// the Header's type is CHAINCODE.  It contains the arguments for this
// invocation.
type PtnChaincodeProposalPayload struct {
	// Input contains the arguments for this invocation. If this invocation
	// deploys a new chaincode, ESCC/VSCC are part of this field.
	// This is usually a marshaled ChaincodeInvocationSpec
	Input []byte `protobuf:"bytes,1,opt,name=input,proto3" json:"input,omitempty"`
	// TransientMap contains data (e.g. cryptographic material) that might be used
	// to implement some form of application-level confidentiality. The contents
	// of this field are supposed to always be omitted from the transaction and
	// excluded from the ledger.
	TransientMap         map[string][]byte `protobuf:"bytes,2,rep,name=TransientMap,proto3" json:"TransientMap,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *PtnChaincodeProposalPayload) Reset()         { *m = PtnChaincodeProposalPayload{} }
func (m *PtnChaincodeProposalPayload) String() string { return proto.CompactTextString(m) }
func (*PtnChaincodeProposalPayload) ProtoMessage()    {}
func (*PtnChaincodeProposalPayload) Descriptor() ([]byte, []int) {
	return fileDescriptor_c3ac5ce23bf32d05, []int{3}
}

func (m *PtnChaincodeProposalPayload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PtnChaincodeProposalPayload.Unmarshal(m, b)
}
func (m *PtnChaincodeProposalPayload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PtnChaincodeProposalPayload.Marshal(b, m, deterministic)
}
func (m *PtnChaincodeProposalPayload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PtnChaincodeProposalPayload.Merge(m, src)
}
func (m *PtnChaincodeProposalPayload) XXX_Size() int {
	return xxx_messageInfo_PtnChaincodeProposalPayload.Size(m)
}
func (m *PtnChaincodeProposalPayload) XXX_DiscardUnknown() {
	xxx_messageInfo_PtnChaincodeProposalPayload.DiscardUnknown(m)
}

var xxx_messageInfo_PtnChaincodeProposalPayload proto.InternalMessageInfo

func (m *PtnChaincodeProposalPayload) GetInput() []byte {
	if m != nil {
		return m.Input
	}
	return nil
}

func (m *PtnChaincodeProposalPayload) GetTransientMap() map[string][]byte {
	if m != nil {
		return m.TransientMap
	}
	return nil
}

// ChaincodeAction contains the actions the events generated by the execution
// of the chaincode.
type PtnChaincodeAction struct {
	// This field contains the read set and the write set produced by the
	// chaincode executing this invocation.
	Results []byte `protobuf:"bytes,1,opt,name=results,proto3" json:"results,omitempty"`
	// This field contains the events generated by the chaincode executing this
	// invocation.
	Events []byte `protobuf:"bytes,2,opt,name=events,proto3" json:"events,omitempty"`
	// This field contains the result of executing this invocation.
	Response *PtnResponse `protobuf:"bytes,3,opt,name=response,proto3" json:"response,omitempty"`
	// This field contains the ChaincodeID of executing this invocation. Endorser
	// will set it with the ChaincodeID called by endorser while simulating proposal.
	// Committer will validate the version matching with latest chaincode version.
	// Adding ChaincodeID to keep version opens up the possibility of multiple
	// ChaincodeAction per transaction.
	ChaincodeId          *PtnChaincodeID `protobuf:"bytes,4,opt,name=chaincode_id,json=chaincodeId,proto3" json:"chaincode_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *PtnChaincodeAction) Reset()         { *m = PtnChaincodeAction{} }
func (m *PtnChaincodeAction) String() string { return proto.CompactTextString(m) }
func (*PtnChaincodeAction) ProtoMessage()    {}
func (*PtnChaincodeAction) Descriptor() ([]byte, []int) {
	return fileDescriptor_c3ac5ce23bf32d05, []int{4}
}

func (m *PtnChaincodeAction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PtnChaincodeAction.Unmarshal(m, b)
}
func (m *PtnChaincodeAction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PtnChaincodeAction.Marshal(b, m, deterministic)
}
func (m *PtnChaincodeAction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PtnChaincodeAction.Merge(m, src)
}
func (m *PtnChaincodeAction) XXX_Size() int {
	return xxx_messageInfo_PtnChaincodeAction.Size(m)
}
func (m *PtnChaincodeAction) XXX_DiscardUnknown() {
	xxx_messageInfo_PtnChaincodeAction.DiscardUnknown(m)
}

var xxx_messageInfo_PtnChaincodeAction proto.InternalMessageInfo

func (m *PtnChaincodeAction) GetResults() []byte {
	if m != nil {
		return m.Results
	}
	return nil
}

func (m *PtnChaincodeAction) GetEvents() []byte {
	if m != nil {
		return m.Events
	}
	return nil
}

func (m *PtnChaincodeAction) GetResponse() *PtnResponse {
	if m != nil {
		return m.Response
	}
	return nil
}

func (m *PtnChaincodeAction) GetChaincodeId() *PtnChaincodeID {
	if m != nil {
		return m.ChaincodeId
	}
	return nil
}

func init() {
	proto.RegisterType((*PtnSignedProposal)(nil), "protos.PtnSignedProposal")
	proto.RegisterType((*PtnProposal)(nil), "protos.PtnProposal")
	proto.RegisterType((*PtnChaincodeHeaderExtension)(nil), "protos.PtnChaincodeHeaderExtension")
	proto.RegisterType((*PtnChaincodeProposalPayload)(nil), "protos.PtnChaincodeProposalPayload")
	proto.RegisterMapType((map[string][]byte)(nil), "protos.PtnChaincodeProposalPayload.TransientMapEntry")
	proto.RegisterType((*PtnChaincodeAction)(nil), "protos.PtnChaincodeAction")
}

func init() { proto.RegisterFile("proposal.proto", fileDescriptor_c3ac5ce23bf32d05) }

var fileDescriptor_c3ac5ce23bf32d05 = []byte{
	// 466 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x93, 0xdd, 0x6a, 0xd4, 0x40,
	0x14, 0xc7, 0xc9, 0xae, 0xb6, 0x76, 0xb2, 0x7e, 0xec, 0x28, 0x25, 0xac, 0x5e, 0x94, 0x80, 0xd0,
	0x9b, 0x6e, 0x60, 0x45, 0x50, 0x6f, 0xd4, 0xad, 0x05, 0x7b, 0x21, 0x84, 0x28, 0x52, 0x05, 0x59,
	0x27, 0xc9, 0x21, 0x3b, 0x34, 0x9d, 0x19, 0x66, 0x4e, 0x16, 0xf7, 0x09, 0x7c, 0x1f, 0x5f, 0xc4,
	0x57, 0x92, 0xec, 0xcc, 0x64, 0x63, 0xeb, 0x85, 0xb0, 0x57, 0xc9, 0xff, 0x9c, 0x33, 0xbf, 0xf3,
	0x35, 0x43, 0xee, 0x29, 0x2d, 0x95, 0x34, 0xac, 0x9e, 0x2a, 0x2d, 0x51, 0xd2, 0xbd, 0xcd, 0xc7,
	0x4c, 0xd2, 0x8a, 0xe3, 0xb2, 0xc9, 0xa7, 0x85, 0xbc, 0x4a, 0x14, 0xab, 0x6b, 0x40, 0x29, 0x20,
	0xa9, 0xe4, 0xc9, 0x56, 0x14, 0x52, 0x43, 0xb2, 0xba, 0x3a, 0x95, 0x02, 0x35, 0x2b, 0x30, 0x6d,
	0xf2, 0xc4, 0x1e, 0x4e, 0x14, 0x80, 0x4e, 0x8a, 0x25, 0xe3, 0xa2, 0x90, 0x25, 0x58, 0xf2, 0xe4,
	0x62, 0x67, 0xa2, 0x2f, 0x75, 0xa1, 0xc1, 0x28, 0x29, 0x8c, 0x23, 0xc7, 0x17, 0x64, 0x9c, 0xa2,
	0xf8, 0xc8, 0x2b, 0x01, 0x65, 0xea, 0x62, 0xe8, 0xd3, 0x6d, 0x6b, 0x8b, 0x7c, 0x8d, 0x60, 0xa2,
	0xe0, 0x28, 0x38, 0x1e, 0x65, 0x77, 0xbd, 0x75, 0xde, 0x1a, 0xe9, 0x13, 0x72, 0x60, 0x78, 0x25,
	0x18, 0x36, 0x1a, 0xa2, 0xc1, 0x26, 0x62, 0x6b, 0x88, 0xbf, 0x91, 0x30, 0x45, 0xd1, 0x31, 0x0f,
	0xc9, 0xde, 0x12, 0x58, 0x09, 0xda, 0xb1, 0x9c, 0xa2, 0x11, 0xd9, 0x57, 0x6c, 0x5d, 0x4b, 0x56,
	0x3a, 0x84, 0x97, 0x2d, 0x1e, 0x7e, 0x20, 0x08, 0xc3, 0xa5, 0x88, 0x86, 0x16, 0xdf, 0x19, 0xe2,
	0x9f, 0x01, 0x79, 0x9c, 0xa2, 0x38, 0xf5, 0x93, 0x7a, 0xbf, 0xc1, 0x9d, 0x79, 0x3f, 0x3d, 0x21,
	0xd4, 0x81, 0x16, 0x2b, 0x6e, 0x78, 0xce, 0x6b, 0x8e, 0x6b, 0x97, 0x7b, 0xec, 0x3c, 0x9f, 0x3b,
	0x07, 0x7d, 0x49, 0x46, 0xdd, 0xd0, 0x17, 0xdc, 0xd6, 0x12, 0xce, 0x0e, 0xed, 0x94, 0xcc, 0xb4,
	0x9f, 0xe9, 0xfc, 0x5d, 0x16, 0x76, 0xb1, 0xe7, 0x65, 0xfc, 0xfb, 0x5a, 0x25, 0xbe, 0xe5, 0xd4,
	0xf5, 0xf1, 0x88, 0xdc, 0xe6, 0x42, 0x35, 0xe8, 0x92, 0x5b, 0x41, 0xbf, 0x90, 0xd1, 0x27, 0xcd,
	0x84, 0xe1, 0x20, 0xf0, 0x03, 0x53, 0xd1, 0xe0, 0x68, 0x78, 0x1c, 0xce, 0x9e, 0xff, 0x2b, 0xe1,
	0x35, 0xe0, 0xb4, 0x7f, 0xee, 0x4c, 0xa0, 0x5e, 0x67, 0x7f, 0xa1, 0x26, 0xaf, 0xc9, 0xf8, 0x46,
	0x08, 0x7d, 0x40, 0x86, 0x97, 0x60, 0x07, 0x70, 0x90, 0xb5, 0xbf, 0x6d, 0x5d, 0x2b, 0x56, 0x37,
	0x7e, 0x75, 0x56, 0xbc, 0x1a, 0xbc, 0x08, 0xe2, 0x5f, 0x01, 0xa1, 0xfd, 0x02, 0xde, 0x16, 0xd8,
	0x8e, 0x34, 0x22, 0xfb, 0x1a, 0x4c, 0x53, 0xa3, 0xbf, 0x0f, 0x5e, 0xb6, 0xcb, 0x85, 0x15, 0x08,
	0x34, 0x8e, 0xe5, 0x14, 0x4d, 0xc8, 0x1d, 0x7f, 0xdf, 0x36, 0x1b, 0x0c, 0x67, 0x0f, 0x7b, 0x0d,
	0x66, 0xce, 0x95, 0x75, 0x41, 0x37, 0xd6, 0x70, 0xeb, 0xbf, 0xd7, 0x30, 0xff, 0x4e, 0x42, 0x17,
	0xd5, 0x5e, 0xf8, 0xf9, 0xfd, 0xed, 0xd4, 0x8a, 0x4b, 0x56, 0xc1, 0xd7, 0x37, 0xbb, 0xbe, 0xa1,
	0xdc, 0xbe, 0xef, 0x67, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff, 0x43, 0xcd, 0x75, 0x08, 0xf8, 0x03,
	0x00, 0x00,
}
