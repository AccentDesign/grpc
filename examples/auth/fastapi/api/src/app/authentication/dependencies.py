from typing import Annotated

import grpc
from fastapi import Depends
from fastapi.security import (
    OAuth2PasswordBearer,
    OAuth2PasswordRequestForm,
    SecurityScopes,
)
from google.protobuf.json_format import MessageToDict

from app.authentication.exceptions import Forbidden, Unauthorized
from app.authentication.schemas import UserRead
from app.grpc import grpc_clients
from protos import auth_pb2

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/auth/token/login")

Oauth2Form = Annotated[OAuth2PasswordRequestForm, Depends()]
Token = Annotated[str, Depends(oauth2_scheme)]


async def current_user(token: Token) -> UserRead:
    try:
        request = auth_pb2.Token(token=token)
        user = await grpc_clients["auth"].User(request, timeout=5)
        data = MessageToDict(
            user,
            including_default_value_fields=True,
            preserving_proto_field_name=True,
        )
        return UserRead(**data)
    except grpc.aio.AioRpcError as e:
        raise Unauthorized() from e


CurrentUser = Annotated[UserRead, Depends(current_user)]


async def current_active_user(
    user: CurrentUser,
    security_scopes: SecurityScopes,
) -> UserRead:
    if not user.is_active:
        raise Unauthorized()
    for scope in security_scopes.scopes:
        if scope not in user.user_type.scopes:
            raise Forbidden()
    return user


CurrentActiveUser = Annotated[UserRead, Depends(current_active_user)]
