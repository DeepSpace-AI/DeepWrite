from datetime import datetime

from worker.celery_app import celery_app


@celery_app.task(name="worker.ping")
def ping(payload: str = "pong") -> dict[str, str]:
    return {"message": payload, "time": datetime.now().isoformat()}
