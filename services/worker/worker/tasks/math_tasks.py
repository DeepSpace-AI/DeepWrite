from worker.celery_app import celery_app


@celery_app.task(name="worker.add")
def add(a: int, b: int) -> int:
    return a + b
