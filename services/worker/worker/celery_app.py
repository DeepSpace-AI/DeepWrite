from celery import Celery

from config.config import get_config

cfg = get_config()

celery_app = Celery(
    "deepwrite_worker",
    broker=cfg.celery_broker_url,
    backend=cfg.celery_result_backend,
    include=["worker.tasks"],
)

celery_app.conf.update(
    task_default_queue=cfg.celery_default_queue,
    task_track_started=True,
    task_acks_late=True,
    worker_prefetch_multiplier=1,
    timezone="Asia/Shanghai",
    enable_utc=False,
)

celery_app.autodiscover_tasks(packages=["worker"], related_name="tasks")
