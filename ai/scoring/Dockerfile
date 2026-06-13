FROM python:3.11-slim

WORKDIR /app

COPY vendor/ /vendor/
COPY scoring/pyproject.toml .

RUN pip install --no-index --find-links /vendor \
    fastapi "uvicorn[standard]" pydantic redis psycopg2-binary python-dotenv numpy

COPY scoring/ .

EXPOSE 9001

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "9001"]
