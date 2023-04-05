from fastapi import HTTPException, status


class BadRequest(HTTPException):
    def __init__(self, detail):
        super().__init__(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=detail,
        )


class IncorrectLoginCredentials(BadRequest):
    def __init__(self):
        super().__init__("INCORRECT_LOGIN_CREDENTIALS")


class Unauthorized(HTTPException):
    def __init__(self):
        super().__init__(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="INVALID_AUTH_CREDENTIALS",
            headers={"WWW-Authenticate": "Bearer"},
        )


class Forbidden(HTTPException):
    def __init__(self):
        super().__init__(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="FORBIDDEN",
        )
