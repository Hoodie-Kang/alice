// Copyright © 2020 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.21.12
// source: github.com/getamis/alice/crypto/homo/paillier/message.proto

package paillier

import (
	zkproof "github.com/getamis/alice/crypto/zkproof"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PubKeyMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Proof *zkproof.IntegerFactorizationProofMessage `protobuf:"bytes,1,opt,name=proof,proto3" json:"proof,omitempty"`
	G     []byte                                    `protobuf:"bytes,2,opt,name=g,proto3" json:"g,omitempty"`
}

func (x *PubKeyMessage) Reset() {
	*x = PubKeyMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_getamis_alice_crypto_homo_paillier_message_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PubKeyMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PubKeyMessage) ProtoMessage() {}

func (x *PubKeyMessage) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_getamis_alice_crypto_homo_paillier_message_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PubKeyMessage.ProtoReflect.Descriptor instead.
func (*PubKeyMessage) Descriptor() ([]byte, []int) {
	return file_github_com_getamis_alice_crypto_homo_paillier_message_proto_rawDescGZIP(), []int{0}
}

func (x *PubKeyMessage) GetProof() *zkproof.IntegerFactorizationProofMessage {
	if x != nil {
		return x.Proof
	}
	return nil
}

func (x *PubKeyMessage) GetG() []byte {
	if x != nil {
		return x.G
	}
	return nil
}

type ZkBetaAndBMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProofB    *zkproof.SchnorrProofMessage `protobuf:"bytes,1,opt,name=proofB,proto3" json:"proofB,omitempty"`
	ProofBeta *zkproof.SchnorrProofMessage `protobuf:"bytes,2,opt,name=proofBeta,proto3" json:"proofBeta,omitempty"`
}

func (x *ZkBetaAndBMessage) Reset() {
	*x = ZkBetaAndBMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_getamis_alice_crypto_homo_paillier_message_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ZkBetaAndBMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ZkBetaAndBMessage) ProtoMessage() {}

func (x *ZkBetaAndBMessage) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_getamis_alice_crypto_homo_paillier_message_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ZkBetaAndBMessage.ProtoReflect.Descriptor instead.
func (*ZkBetaAndBMessage) Descriptor() ([]byte, []int) {
	return file_github_com_getamis_alice_crypto_homo_paillier_message_proto_rawDescGZIP(), []int{1}
}

func (x *ZkBetaAndBMessage) GetProofB() *zkproof.SchnorrProofMessage {
	if x != nil {
		return x.ProofB
	}
	return nil
}

func (x *ZkBetaAndBMessage) GetProofBeta() *zkproof.SchnorrProofMessage {
	if x != nil {
		return x.ProofBeta
	}
	return nil
}

var File_github_com_getamis_alice_crypto_homo_paillier_message_proto protoreflect.FileDescriptor

var file_github_com_getamis_alice_crypto_homo_paillier_message_proto_rawDesc = []byte{
	0x0a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x65, 0x74,
	0x61, 0x6d, 0x69, 0x73, 0x2f, 0x61, 0x6c, 0x69, 0x63, 0x65, 0x2f, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x6f, 0x2f, 0x68, 0x6f, 0x6d, 0x6f, 0x2f, 0x70, 0x61, 0x69, 0x6c, 0x6c, 0x69, 0x65, 0x72, 0x2f,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x22, 0x67,
	0x65, 0x74, 0x61, 0x6d, 0x69, 0x73, 0x2e, 0x61, 0x6c, 0x69, 0x63, 0x65, 0x2e, 0x63, 0x72, 0x79,
	0x70, 0x74, 0x6f, 0x2e, 0x68, 0x6f, 0x6d, 0x6f, 0x2e, 0x70, 0x61, 0x69, 0x6c, 0x6c, 0x69, 0x65,
	0x72, 0x1a, 0x35, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x65,
	0x74, 0x61, 0x6d, 0x69, 0x73, 0x2f, 0x61, 0x6c, 0x69, 0x63, 0x65, 0x2f, 0x63, 0x72, 0x79, 0x70,
	0x74, 0x6f, 0x2f, 0x7a, 0x6b, 0x70, 0x72, 0x6f, 0x6f, 0x66, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x73, 0x0a, 0x0d, 0x50, 0x75, 0x62, 0x4b,
	0x65, 0x79, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x54, 0x0a, 0x05, 0x70, 0x72, 0x6f,
	0x6f, 0x66, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x3e, 0x2e, 0x67, 0x65, 0x74, 0x61, 0x6d,
	0x69, 0x73, 0x2e, 0x61, 0x6c, 0x69, 0x63, 0x65, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x2e,
	0x7a, 0x6b, 0x70, 0x72, 0x6f, 0x6f, 0x66, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x65, 0x72, 0x46,
	0x61, 0x63, 0x74, 0x6f, 0x72, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x6f,
	0x66, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x05, 0x70, 0x72, 0x6f, 0x6f, 0x66, 0x12,
	0x0c, 0x0a, 0x01, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x01, 0x67, 0x22, 0xaf, 0x01,
	0x0a, 0x11, 0x7a, 0x6b, 0x42, 0x65, 0x74, 0x61, 0x41, 0x6e, 0x64, 0x42, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x49, 0x0a, 0x06, 0x70, 0x72, 0x6f, 0x6f, 0x66, 0x42, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x31, 0x2e, 0x67, 0x65, 0x74, 0x61, 0x6d, 0x69, 0x73, 0x2e, 0x61, 0x6c,
	0x69, 0x63, 0x65, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x2e, 0x7a, 0x6b, 0x70, 0x72, 0x6f,
	0x6f, 0x66, 0x2e, 0x53, 0x63, 0x68, 0x6e, 0x6f, 0x72, 0x72, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x06, 0x70, 0x72, 0x6f, 0x6f, 0x66, 0x42, 0x12, 0x4f,
	0x0a, 0x09, 0x70, 0x72, 0x6f, 0x6f, 0x66, 0x42, 0x65, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x31, 0x2e, 0x67, 0x65, 0x74, 0x61, 0x6d, 0x69, 0x73, 0x2e, 0x61, 0x6c, 0x69, 0x63,
	0x65, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x2e, 0x7a, 0x6b, 0x70, 0x72, 0x6f, 0x6f, 0x66,
	0x2e, 0x53, 0x63, 0x68, 0x6e, 0x6f, 0x72, 0x72, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6f, 0x66, 0x42, 0x65, 0x74, 0x61, 0x42,
	0x2f, 0x5a, 0x2d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x65,
	0x74, 0x61, 0x6d, 0x69, 0x73, 0x2f, 0x61, 0x6c, 0x69, 0x63, 0x65, 0x2f, 0x63, 0x72, 0x79, 0x70,
	0x74, 0x6f, 0x2f, 0x68, 0x6f, 0x6d, 0x6f, 0x2f, 0x70, 0x61, 0x69, 0x6c, 0x6c, 0x69, 0x65, 0x72,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_getamis_alice_crypto_homo_paillier_message_proto_rawDescOnce sync.Once
	file_github_com_getamis_alice_crypto_homo_paillier_message_proto_rawDescData = file_github_com_getamis_alice_crypto_homo_paillier_message_proto_rawDesc
)

func file_github_com_getamis_alice_crypto_homo_paillier_message_proto_rawDescGZIP() []byte {
	file_github_com_getamis_alice_crypto_homo_paillier_message_proto_rawDescOnce.Do(func() {
		file_github_com_getamis_alice_crypto_homo_paillier_message_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_getamis_alice_crypto_homo_paillier_message_proto_rawDescData)
	})
	return file_github_com_getamis_alice_crypto_homo_paillier_message_proto_rawDescData
}

var file_github_com_getamis_alice_crypto_homo_paillier_message_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_github_com_getamis_alice_crypto_homo_paillier_message_proto_goTypes = []interface{}{
	(*PubKeyMessage)(nil),                            // 0: getamis.alice.crypto.homo.paillier.PubKeyMessage
	(*ZkBetaAndBMessage)(nil),                        // 1: getamis.alice.crypto.homo.paillier.zkBetaAndBMessage
	(*zkproof.IntegerFactorizationProofMessage)(nil), // 2: getamis.alice.crypto.zkproof.IntegerFactorizationProofMessage
	(*zkproof.SchnorrProofMessage)(nil),              // 3: getamis.alice.crypto.zkproof.SchnorrProofMessage
}
var file_github_com_getamis_alice_crypto_homo_paillier_message_proto_depIdxs = []int32{
	2, // 0: getamis.alice.crypto.homo.paillier.PubKeyMessage.proof:type_name -> getamis.alice.crypto.zkproof.IntegerFactorizationProofMessage
	3, // 1: getamis.alice.crypto.homo.paillier.zkBetaAndBMessage.proofB:type_name -> getamis.alice.crypto.zkproof.SchnorrProofMessage
	3, // 2: getamis.alice.crypto.homo.paillier.zkBetaAndBMessage.proofBeta:type_name -> getamis.alice.crypto.zkproof.SchnorrProofMessage
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_github_com_getamis_alice_crypto_homo_paillier_message_proto_init() }
func file_github_com_getamis_alice_crypto_homo_paillier_message_proto_init() {
	if File_github_com_getamis_alice_crypto_homo_paillier_message_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_getamis_alice_crypto_homo_paillier_message_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PubKeyMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_github_com_getamis_alice_crypto_homo_paillier_message_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ZkBetaAndBMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_getamis_alice_crypto_homo_paillier_message_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_getamis_alice_crypto_homo_paillier_message_proto_goTypes,
		DependencyIndexes: file_github_com_getamis_alice_crypto_homo_paillier_message_proto_depIdxs,
		MessageInfos:      file_github_com_getamis_alice_crypto_homo_paillier_message_proto_msgTypes,
	}.Build()
	File_github_com_getamis_alice_crypto_homo_paillier_message_proto = out.File
	file_github_com_getamis_alice_crypto_homo_paillier_message_proto_rawDesc = nil
	file_github_com_getamis_alice_crypto_homo_paillier_message_proto_goTypes = nil
	file_github_com_getamis_alice_crypto_homo_paillier_message_proto_depIdxs = nil
}
