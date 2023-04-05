from queue import Queue
from typing import Generic, Type, TypedDict, TypeVar

import grpc
from grpc.aio import Channel

from app.config import settings
from protos.auth_pb2_grpc import AuthenticationStub

StubType = TypeVar("StubType")


class GrpcClientPool(Generic[StubType]):
    def __init__(
        self,
        stub_class: Type[StubType],
        host: str,
        port: int,
        pool_size: int = 10,
    ):
        self.stub_class: Type[StubType] = stub_class
        self.host: str = host
        self.port: int = port
        self.pool_size: int = pool_size
        self.pool = Queue(maxsize=pool_size)

    async def get_channel(self) -> Channel:
        if self.pool.qsize() == 0:
            return grpc.aio.insecure_channel(f"{self.host}:{self.port}")
        else:
            return self.pool.get_nowait()

    def release_channel(self, channel) -> None:
        if self.pool.qsize() < self.pool_size:
            self.pool.put_nowait(channel)
        else:
            channel.close()

    async def __aenter__(self) -> "GrpcClientPool":
        return self

    async def __aexit__(self, exc_type, exc_value, traceback) -> None:
        while not self.pool.empty():
            channel = self.pool.get_nowait()
            await channel.close()

    async def __call__(self, method: str, *args, **kwargs):
        channel = await self.get_channel()
        try:
            stub = self.stub_class(channel)
            return await getattr(stub, method)(*args, **kwargs)
        finally:
            self.release_channel(channel)


class AuthClient(GrpcClientPool[AuthenticationStub]):
    def __init__(self):
        super().__init__(AuthenticationStub, settings.auth_host, settings.auth_port)


class GrpcClients(TypedDict, total=False):
    auth: AuthClient


grpc_clients: GrpcClients = {}
