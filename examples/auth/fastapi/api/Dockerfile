FROM        python:3.11-slim as base

RUN         pip install --upgrade pip poetry

WORKDIR     /app

COPY        ./src/pyproject.toml ./src/poetry.lock ./

ARG         ENVIRONMENT=production

RUN         poetry config virtualenvs.create false && \
            poetry install $(test "$ENVIRONMENT" = production && echo "--only main") --no-interaction --no-ansi && \
            rm -rf /root/.cache/pypoetry

# Start a new stage to copy the application code
FROM        base as final

ENV         PYTHONDONTWRITEBYTECODE=1
ENV         PYTHONFAULTHANDLER=1
ENV         PYTHONPATH=/app

WORKDIR     /app

COPY        ./src .

EXPOSE      80

CMD         ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "80"]
