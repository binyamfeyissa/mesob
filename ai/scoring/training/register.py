"""Sign and register a trained model artifact."""
import hashlib
import json
import os
import time


def _compute_sha256(path: str) -> str:
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(65536), b""):
            h.update(chunk)
    return h.hexdigest()


def _record_in_db(version: str, sha256: str, metrics: dict) -> None:
    """Write candidate version to scoring_models table. Non-fatal if DB unavailable."""
    try:
        import psycopg2
        url = os.getenv("MESOB_SCORING_DB_URL", "postgresql://mesob:mesob@localhost:5432/mesob_scoring")
        conn = psycopg2.connect(url)
        with conn, conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO scoring_models
                    (version, sha256, auc, disparate_impact, status, registered_at)
                VALUES (%s, %s, %s, %s, 'CANDIDATE', NOW())
                ON CONFLICT (version) DO UPDATE
                    SET sha256 = EXCLUDED.sha256,
                        auc    = EXCLUDED.auc,
                        disparate_impact = EXCLUDED.disparate_impact,
                        status = 'CANDIDATE',
                        registered_at = NOW()
                """,
                (
                    version,
                    sha256,
                    metrics.get("auc", 0.0),
                    metrics.get("disparate_impact", 1.0),
                ),
            )
        conn.close()
    except Exception as e:
        print(f"[register] DB unavailable — skipping DB record: {e}")


def register(model_path: str, metrics: dict) -> str:
    """Sign model artifact and push to model registry. Returns version string."""
    if not os.path.exists(model_path):
        raise FileNotFoundError(f"Model artifact not found: {model_path}")

    if not metrics.get("passed", False):
        raise ValueError("Model failed evaluation gates — refusing to register")

    sha256 = _compute_sha256(model_path)
    version = f"v{int(time.time())}"

    # Write manifest alongside the model file.
    manifest = {
        "version": version,
        "sha256": sha256,
        "auc": metrics.get("auc"),
        "disparate_impact": metrics.get("disparate_impact"),
        "model_path": os.path.abspath(model_path),
    }
    manifest_path = model_path.replace(".pkl", f"_{version}.manifest.json")
    with open(manifest_path, "w") as f:
        json.dump(manifest, f, indent=2)

    _record_in_db(version, sha256, metrics)

    print(f"Registered model {version} (sha256={sha256[:12]}…)")
    print(f"Manifest: {manifest_path}")
    return version
