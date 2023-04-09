import contextlib

import grpc
from fastapi import APIRouter, status
from google.protobuf.json_format import MessageToDict

from app.authentication.dependencies import (
    CurrentActiveUser,
    CurrentUser,
    Oauth2Form,
    Token,
)
from app.authentication.exceptions import BadRequest, IncorrectLoginCredentials
from app.authentication.schemas import (
    BearerResponse,
    ForgotPassword,
    ResetPassword,
    UserCreate,
    UserRead,
    UserUpdate,
    VerifyRequest,
    VerifyToken,
)
from app.grpc import grpc_clients
from protos import auth_pb2

router = APIRouter()


@router.post(
    "/auth/token/login",
    response_model=BearerResponse,
    tags=["auth"],
)
async def login(data: Oauth2Form) -> BearerResponse:
    try:
        request = auth_pb2.BearerTokenRequest(
            email=data.username,
            password=data.password,
        )
        response = await grpc_clients["auth"].BearerToken(request, timeout=5)
        return MessageToDict(response, preserving_proto_field_name=True)
    except grpc.aio.AioRpcError as e:
        raise IncorrectLoginCredentials() from e


@router.post("/auth/token/logout", tags=["auth"])
async def logout(_: CurrentUser, token: Token) -> None:
    async with grpc_clients["auth"] as client:
        with contextlib.suppress(grpc.aio.AioRpcError):
            request = auth_pb2.Token(token=token)
            await client.RevokeBearerToken(request, timeout=5)


@router.post(
    "/auth/register",
    response_model=UserRead,
    status_code=status.HTTP_201_CREATED,
    tags=["auth"],
)
async def register(data: UserCreate) -> UserRead:
    try:
        request = auth_pb2.RegisterRequest(**data.dict())
        response = await grpc_clients["auth"].Register(request, timeout=5)
        return MessageToDict(
            response,
            preserving_proto_field_name=True,
            including_default_value_fields=True,
        )
    except grpc.aio.AioRpcError as e:
        raise BadRequest(e.details()) from e


@router.post(
    "/auth/verify-request",
    status_code=status.HTTP_202_ACCEPTED,
    tags=["auth"],
)
async def verify_request(data: VerifyRequest) -> None:
    with contextlib.suppress(grpc.aio.AioRpcError):
        request = auth_pb2.VerifyUserTokenRequest(email=data.email)
        response = await grpc_clients["auth"].VerifyUserToken(request, timeout=5)

        print("Success: VerifyUserToken")
        print("-" * 60)
        print(response)

        # add a background task to send an email with the details
        # DO NOT expose the token or the success/failure


@router.post(
    "/auth/verify",
    response_model=UserRead,
    tags=["auth"],
)
async def verify(data: VerifyToken) -> UserRead:
    try:
        request = auth_pb2.Token(token=data.token)
        response = await grpc_clients["auth"].VerifyUser(request, timeout=5)
        return MessageToDict(
            response,
            preserving_proto_field_name=True,
            including_default_value_fields=True,
        )
    except grpc.aio.AioRpcError as e:
        raise BadRequest(e.details()) from e


@router.post(
    "/auth/forgot-password",
    status_code=status.HTTP_202_ACCEPTED,
    tags=["auth"],
)
async def forgot_password(data: ForgotPassword) -> None:
    with contextlib.suppress(grpc.aio.AioRpcError):
        request = auth_pb2.ResetPasswordTokenRequest(email=data.email)
        response = await grpc_clients["auth"].ResetPasswordToken(request, timeout=5)

        print("Success: ResetPasswordToken")
        print("-" * 60)
        print(response)

        # add a background task to send an email with the details
        # DO NOT expose the token or the success/failure


@router.post(
    "/auth/reset-password",
    tags=["auth"],
)
async def reset_password(data: ResetPassword) -> None:
    try:
        request = auth_pb2.ResetPasswordRequest(
            token=data.token,
            password=data.password,
        )
        await grpc_clients["auth"].ResetPassword(request, timeout=5)
    except grpc.aio.AioRpcError as e:
        raise BadRequest(e.details()) from e


@router.get(
    "/users/me",
    response_model=UserRead,
    tags=["users"],
)
async def get_current_user(user: CurrentActiveUser) -> UserRead:
    return user


@router.patch(
    "/users/me",
    response_model=UserRead,
    tags=["users"],
)
async def update_current_user(
    data: UserUpdate,
    token: Token,
    _: CurrentActiveUser,
) -> UserRead:
    try:
        request = auth_pb2.UpdateUserRequest(
            token=token, **data.dict(exclude_unset=True)
        )
        response = await grpc_clients["auth"].UpdateUser(request, timeout=5)
        return MessageToDict(
            response,
            preserving_proto_field_name=True,
            including_default_value_fields=True,
        )
    except grpc.aio.AioRpcError as e:
        raise BadRequest(e.details()) from e
