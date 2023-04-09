from contextlib import asynccontextmanager
from typing import Annotated

from fastapi import FastAPI, Security
from fastapi.middleware.cors import CORSMiddleware
from starlette.middleware import Middleware
from starlette.middleware.trustedhost import TrustedHostMiddleware

from app.authentication.dependencies import current_active_user
from app.authentication.routes import router
from app.authentication.schemas import UserRead
from app.config import settings
from app.grpc import AuthGrpcClient, grpc_clients


@asynccontextmanager
async def lifespan(app: FastAPI):
    async with AuthGrpcClient(settings.auth_host, settings.auth_port) as client:
        grpc_clients["auth"] = client
        yield
    grpc_clients.clear()


middleware = [
    Middleware(
        TrustedHostMiddleware,
        allowed_hosts=settings.allowed_hosts,
    ),
    Middleware(
        CORSMiddleware,
        allow_origins=settings.allow_origins,
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    ),
]


app = FastAPI(
    middleware=middleware,
    lifespan=lifespan,
)


@app.get("/")
async def root() -> str:
    return "ok"


@app.get("/read", response_model=UserRead)
async def read(
    current_user: Annotated[UserRead, Security(current_active_user, scopes=["read"])],
) -> UserRead:
    return current_user


@app.get("/write", response_model=UserRead)
async def write(
    current_user: Annotated[UserRead, Security(current_active_user, scopes=["write"])],
) -> UserRead:
    return current_user


@app.get("/admin", response_model=UserRead)
async def admin(
    current_user: Annotated[UserRead, Security(current_active_user, scopes=["admin"])],
) -> UserRead:
    return current_user


app.include_router(router)
