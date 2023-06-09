# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: auth.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\nauth.proto\x12\x08pkg.auth\"\x07\n\x05\x45mpty\"\x16\n\x05Token\x12\r\n\x05token\x18\x01 \x01(\t\"5\n\x12\x42\x65\x61rerTokenRequest\x12\r\n\x05\x65mail\x18\x01 \x01(\t\x12\x10\n\x08password\x18\x02 \x01(\t\"O\n\x13\x42\x65\x61rerTokenResponse\x12\x14\n\x0c\x61\x63\x63\x65ss_token\x18\x01 \x01(\t\x12\x12\n\ntoken_type\x18\x02 \x01(\t\x12\x0e\n\x06\x65xpiry\x18\x03 \x01(\x05\"Y\n\x0fRegisterRequest\x12\r\n\x05\x65mail\x18\x01 \x01(\t\x12\x10\n\x08password\x18\x02 \x01(\t\x12\x12\n\nfirst_name\x18\x03 \x01(\t\x12\x11\n\tlast_name\x18\x04 \x01(\t\"7\n\x14ResetPasswordRequest\x12\r\n\x05token\x18\x01 \x01(\t\x12\x10\n\x08password\x18\x02 \x01(\t\"*\n\x19ResetPasswordTokenRequest\x12\r\n\x05\x65mail\x18\x01 \x01(\t\"j\n\x11UpdateUserRequest\x12\r\n\x05token\x18\x01 \x01(\t\x12\r\n\x05\x65mail\x18\x02 \x01(\t\x12\x10\n\x08password\x18\x03 \x01(\t\x12\x12\n\nfirst_name\x18\x04 \x01(\t\x12\x11\n\tlast_name\x18\x05 \x01(\t\"(\n\x08UserType\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x0e\n\x06scopes\x18\x02 \x03(\t\"\x9f\x01\n\x0cUserResponse\x12\n\n\x02id\x18\x01 \x01(\t\x12\r\n\x05\x65mail\x18\x02 \x01(\t\x12\x12\n\nfirst_name\x18\x03 \x01(\t\x12\x11\n\tlast_name\x18\x04 \x01(\t\x12%\n\tuser_type\x18\x05 \x01(\x0b\x32\x12.pkg.auth.UserType\x12\x11\n\tis_active\x18\x06 \x01(\x08\x12\x13\n\x0bis_verified\x18\x07 \x01(\x08\"\'\n\x16VerifyUserTokenRequest\x12\r\n\x05\x65mail\x18\x01 \x01(\t\"U\n\x0eTokenWithEmail\x12\r\n\x05token\x18\x01 \x01(\t\x12\r\n\x05\x65mail\x18\x02 \x01(\t\x12\x12\n\nfirst_name\x18\x03 \x01(\t\x12\x11\n\tlast_name\x18\x04 \x01(\t2\xf5\x04\n\x0e\x41uthentication\x12L\n\x0b\x42\x65\x61rerToken\x12\x1c.pkg.auth.BearerTokenRequest\x1a\x1d.pkg.auth.BearerTokenResponse\"\x00\x12\x37\n\x11RevokeBearerToken\x12\x0f.pkg.auth.Token\x1a\x0f.pkg.auth.Empty\"\x00\x12?\n\x08Register\x12\x19.pkg.auth.RegisterRequest\x1a\x16.pkg.auth.UserResponse\"\x00\x12\x42\n\rResetPassword\x12\x1e.pkg.auth.ResetPasswordRequest\x1a\x0f.pkg.auth.Empty\"\x00\x12U\n\x12ResetPasswordToken\x12#.pkg.auth.ResetPasswordTokenRequest\x1a\x18.pkg.auth.TokenWithEmail\"\x00\x12\x31\n\x04User\x12\x0f.pkg.auth.Token\x1a\x16.pkg.auth.UserResponse\"\x00\x12\x43\n\nUpdateUser\x12\x1b.pkg.auth.UpdateUserRequest\x1a\x16.pkg.auth.UserResponse\"\x00\x12\x37\n\nVerifyUser\x12\x0f.pkg.auth.Token\x1a\x16.pkg.auth.UserResponse\"\x00\x12O\n\x0fVerifyUserToken\x12 .pkg.auth.VerifyUserTokenRequest\x1a\x18.pkg.auth.TokenWithEmail\"\x00\x42\x30Z.github.com/accentdesign/grpc/services/auth/pkgb\x06proto3')

_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, globals())
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'auth_pb2', globals())
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z.github.com/accentdesign/grpc/services/auth/pkg'
  _EMPTY._serialized_start=24
  _EMPTY._serialized_end=31
  _TOKEN._serialized_start=33
  _TOKEN._serialized_end=55
  _BEARERTOKENREQUEST._serialized_start=57
  _BEARERTOKENREQUEST._serialized_end=110
  _BEARERTOKENRESPONSE._serialized_start=112
  _BEARERTOKENRESPONSE._serialized_end=191
  _REGISTERREQUEST._serialized_start=193
  _REGISTERREQUEST._serialized_end=282
  _RESETPASSWORDREQUEST._serialized_start=284
  _RESETPASSWORDREQUEST._serialized_end=339
  _RESETPASSWORDTOKENREQUEST._serialized_start=341
  _RESETPASSWORDTOKENREQUEST._serialized_end=383
  _UPDATEUSERREQUEST._serialized_start=385
  _UPDATEUSERREQUEST._serialized_end=491
  _USERTYPE._serialized_start=493
  _USERTYPE._serialized_end=533
  _USERRESPONSE._serialized_start=536
  _USERRESPONSE._serialized_end=695
  _VERIFYUSERTOKENREQUEST._serialized_start=697
  _VERIFYUSERTOKENREQUEST._serialized_end=736
  _TOKENWITHEMAIL._serialized_start=738
  _TOKENWITHEMAIL._serialized_end=823
  _AUTHENTICATION._serialized_start=826
  _AUTHENTICATION._serialized_end=1455
# @@protoc_insertion_point(module_scope)
