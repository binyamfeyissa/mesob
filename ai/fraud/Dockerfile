FROM python:3.11-slim

WORKDIR /app

COPY vendor/ /vendor/
COPY fraud/pyproject.toml .

RUN pip install --no-index --find-links /vendor \
    fastapi "uvicorn[standard]" pydantic redis psycopg2-binary python-dotenv numpy

COPY fraud/ .

EXPOSE 9002

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "9002"]
