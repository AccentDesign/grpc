from typing import List

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    allow_origins: List[str] = ["*"]
    allowed_hosts: List[str] = ["*"]
    auth_host: str = "auth"
    auth_port: int = 50051


settings = Settings()
