import uuid
from typing import List, Optional

from pydantic import BaseModel, EmailStr, constr


class BearerResponse(BaseModel):
    access_token: str
    token_type: str
    expiry: int


class ForgotPassword(BaseModel):
    email: EmailStr


class ResetPassword(BaseModel):
    token: str
    password: constr(min_length=6, max_length=79)


class UserCreate(BaseModel):
    email: EmailStr
    first_name: constr(min_length=1, max_length=120)
    last_name: constr(min_length=1, max_length=120)
    password: constr(min_length=6, max_length=79)


class UserType(BaseModel):
    name: str
    scopes: List[str]


class UserRead(BaseModel):
    id: uuid.UUID
    email: EmailStr
    first_name: constr(min_length=1, max_length=120)
    last_name: constr(min_length=1, max_length=120)
    user_type: UserType
    is_active: bool
    is_verified: bool


class UserUpdate(BaseModel):
    email: Optional[EmailStr]
    first_name: Optional[constr(min_length=1, max_length=120)]
    last_name: Optional[constr(min_length=1, max_length=120)]
    password: Optional[constr(min_length=6, max_length=79)]


class VerifyRequest(BaseModel):
    email: EmailStr


class VerifyToken(BaseModel):
    token: str
